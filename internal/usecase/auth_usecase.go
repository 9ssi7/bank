package usecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/pkg/agent"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/state"
	"github.com/9ssi7/bank/pkg/token"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type UserRepo interface {
	FindByEmail(ctx context.Context, trc trace.Tracer, email string) (*user.User, error)
	FindById(ctx context.Context, trc trace.Tracer, id uuid.UUID) (*user.User, error)
	FindByToken(ctx context.Context, trc trace.Tracer, token string) (*user.User, error)
	IsExistsByEmail(ctx context.Context, trc trace.Tracer, email string) (bool, error)
	Save(ctx context.Context, trc trace.Tracer, user *user.User) error
}

type VerifyRepo interface {
	Save(ctx context.Context, trc trace.Tracer, token string, verify *auth.Verify) error
	IsExists(ctx context.Context, trc trace.Tracer, token string, deviceId string) (bool, error)
	Find(ctx context.Context, trc trace.Tracer, token string, deviceId string) (*auth.Verify, error)
	Delete(ctx context.Context, trc trace.Tracer, token string, deviceId string) error
}

type SessionRepo interface {
	Save(ctx context.Context, trc trace.Tracer, userId uuid.UUID, session *auth.Session) error
	FindByIds(ctx context.Context, trc trace.Tracer, userId uuid.UUID, deviceId string) (*auth.Session, bool, error)
}

type TokenSrv interface {
	GenerateAccessToken(ctx context.Context, u token.User) (string, error)
	GenerateRefreshToken(ctx context.Context, u token.User) (string, error)
	Parse(ctx context.Context, token string) (*token.UserClaim, error)
	Verify(ctx context.Context, token string) (bool, error)
	VerifyAndParse(ctx context.Context, token string) (*token.UserClaim, error)
}

type EventSrv interface {
	Publish(ctx context.Context, sub string, data interface{}) error
}

type AuthUseCase struct {
	tokenSrv    TokenSrv
	eventSrv    EventSrv
	verifyRepo  VerifyRepo
	userRepo    UserRepo
	sessionRepo SessionRepo
}

type AuthLoginVerifyCheckOptions struct {
	VerifyToken string
}

func (u *AuthUseCase) LoginVerifyCheck(ctx context.Context, trc trace.Tracer, opts AuthLoginVerifyCheckOptions) error {
	ctx, span := trc.Start(ctx, "AuthUseCase.VerifyCheck")
	defer span.End()
	exists, err := u.verifyRepo.IsExists(ctx, trc, opts.VerifyToken, state.GetDeviceId(ctx))
	if err != nil {
		return err
	}
	if !exists {
		return rescode.NotFound(errors.New("verify token not exists"))
	}
	return nil
}

type AuthLoginStartOptions struct {
	Email  string
	Device agent.Device
}

func (u *AuthUseCase) LoginStart(ctx context.Context, trc trace.Tracer, opts AuthLoginStartOptions) (*string, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.LoginStart")
	defer span.End()
	usr, err := u.userRepo.FindByEmail(ctx, trc, opts.Email)
	if err != nil {
		return nil, err
	}
	if usr == nil {
		return nil, user.NotFound(errors.New("user not found"))
	}
	if !usr.IsActive {
		return nil, user.Disabled(errors.New("user disabled"))
	}
	if usr.TempToken != nil && *usr.TempToken != "" {
		return nil, user.VerifyRequired(errors.New("user verify required"))
	}
	verifyToken := uuid.New().String()
	verify := auth.NewVerify(auth.VerifyConfig{
		UserId:   usr.Id,
		DeviceId: state.GetDeviceId(ctx),
		Locale:   state.GetLocale(ctx),
	})
	if err := u.verifyRepo.Save(ctx, trc, verifyToken, verify); err != nil {
		return nil, err
	}
	err = u.eventSrv.Publish(ctx, auth.SubjectLoginStarted, &auth.EventLoginStarted{
		Email:  usr.Email,
		Code:   verify.Code,
		Device: opts.Device,
	})
	if err != nil {
		return nil, err
	}
	return &verifyToken, nil
}

type AuthLoginVerifyOptions struct {
	Code        string
	VerifyToken string
	Device      agent.Device
}

func (u *AuthUseCase) LoginVerify(ctx context.Context, trc trace.Tracer, opts AuthLoginVerifyOptions) (*string, *string, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.LoginVerify")
	defer span.End()
	verify, err := u.verifyRepo.Find(ctx, trc, opts.VerifyToken, state.GetDeviceId(ctx))
	if err != nil {
		return nil, nil, err
	}
	if verify.IsExpired() {
		return nil, nil, auth.VerificationExpired(errors.New("verification expired"))
	}
	if verify.IsExceeded() {
		return nil, nil, auth.VerificationExceeded(errors.New("verification exceeded"))
	}
	if opts.Code != verify.Code {
		verify.IncTryCount()
		err = u.verifyRepo.Save(ctx, trc, opts.VerifyToken, verify)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, auth.VerificationInvalid(errors.New("verification invalid"))
	}
	err = u.verifyRepo.Delete(ctx, trc, opts.VerifyToken, state.GetDeviceId(ctx))
	if err != nil {
		return nil, nil, err
	}
	user, err := u.userRepo.FindById(ctx, trc, verify.UserId)
	if err != nil {
		return nil, nil, err
	}
	claims := token.User{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
	accessToken, err := u.tokenSrv.GenerateAccessToken(ctx, claims)
	if err != nil {
		return nil, nil, rescode.Failed(err)
	}
	refreshToken, err := u.tokenSrv.GenerateRefreshToken(ctx, claims)
	if err != nil {
		return nil, nil, rescode.Failed(err)
	}
	ses := auth.NewSession(auth.SessionConfig{
		Device:       opts.Device,
		DeviceId:     state.GetDeviceId(ctx),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
	if err = u.sessionRepo.Save(ctx, trc, user.Id, ses); err != nil {
		return nil, nil, err
	}
	return &accessToken, &refreshToken, nil
}

type AuthRefreshTokenOptions struct {
	UserId     uuid.UUID
	AccessTkn  string
	RefreshTkn string
	IpAddr     string
}

func (u *AuthUseCase) RefreshToken(ctx context.Context, trc trace.Tracer, opts AuthRefreshTokenOptions) (*string, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.RefreshToken")
	defer span.End()
	session, notFound, err := u.sessionRepo.FindByIds(ctx, trc, opts.UserId, state.GetDeviceId(ctx))
	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, auth.InvalidRefreshOrAccessTokens(errors.New("invalid refresh with access token and ip"))
	}
	if !session.IsRefreshValid(opts.AccessTkn, opts.RefreshTkn, opts.IpAddr) {
		return nil, auth.InvalidRefreshOrAccessTokens(errors.New("invalid refresh with access token and ip"))
	}
	user, err := u.userRepo.FindById(ctx, trc, opts.UserId)
	if err != nil {
		return nil, err
	}
	claims := token.User{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
	accessToken, err := u.tokenSrv.GenerateAccessToken(ctx, claims)
	if err != nil {
		return nil, rescode.Failed(err)
	}
	session.Refresh(accessToken)
	if err := u.sessionRepo.Save(ctx, trc, user.Id, session); err != nil {
		return nil, err
	}
	return &accessToken, nil
}

type AuthRegisterOptions struct {
	Name  string
	Email string
}

func (u *AuthUseCase) Register(ctx context.Context, trc trace.Tracer, opts AuthRegisterOptions) error {
	ctx, span := trc.Start(ctx, "AuthUseCase.Register")
	defer span.End()
	exists, err := u.userRepo.IsExistsByEmail(ctx, trc, opts.Email)
	if err != nil {
		return err
	}
	if exists {
		return user.EmailAlreadyExists(errors.New("email already exists"))
	}
	usr := user.New(user.Config{
		Name:  opts.Name,
		Email: opts.Email,
	})
	err = u.userRepo.Save(ctx, trc, usr)
	if err != nil {
		return err
	}
	err = u.eventSrv.Publish(ctx, user.SubjectCreated, &user.EventCreated{
		Name:      opts.Name,
		Email:     opts.Email,
		TempToken: *usr.TempToken,
	})
	if err != nil {
		return err
	}
	return nil
}

type AuthRegistrationVerifyOptions struct {
	Token string
}

func (u *AuthUseCase) RegistrationVerify(ctx context.Context, trc trace.Tracer, opts AuthRegistrationVerifyOptions) error {
	ctx, span := trc.Start(ctx, "AuthUseCase.RegistrationVerify")
	defer span.End()
	usr, err := u.userRepo.FindByToken(ctx, trc, opts.Token)
	if err != nil {
		return err
	}
	usr.Verify()
	return u.userRepo.Save(ctx, trc, usr)
}

type AuthVerifyAccessOptions struct {
	AccessTkn  string
	IpAddr     string
	SkipVerify bool
}

func (u *AuthUseCase) VerifyAccess(ctx context.Context, trc trace.Tracer, opts AuthVerifyAccessOptions) (*token.UserClaim, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.VerifyAccess")
	defer span.End()
	var claims *token.UserClaim
	var err error
	if opts.SkipVerify {
		claims, err = u.tokenSrv.Parse(ctx, opts.AccessTkn)
	} else {
		claims, err = u.tokenSrv.VerifyAndParse(ctx, opts.AccessTkn)
	}
	if err != nil {
		return nil, rescode.Failed(err)
	}
	session, notExists, err := u.sessionRepo.FindByIds(ctx, trc, claims.Id, state.GetDeviceId(ctx))
	if err != nil {
		return nil, err
	}
	if notExists {
		return nil, auth.InvalidAccess(errors.New("invalid access with token and ip"))
	}
	if !session.IsAccessValid(opts.AccessTkn, opts.IpAddr) {
		return nil, auth.InvalidAccess(errors.New("invalid access with token and ip"))
	}
	return claims, nil
}

type AuthVerifyRefreshOptions struct {
	AccessTkn  string
	RefreshTkn string
	IpAddr     string
}

func (u *AuthUseCase) VerifyRefresh(ctx context.Context, trc trace.Tracer, opts *AuthVerifyRefreshOptions) (*token.UserClaim, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.VerifyRefresh")
	defer span.End()
	claims, err := u.tokenSrv.Parse(ctx, opts.RefreshTkn)
	if err != nil {
		return nil, rescode.Failed(err)
	}
	isValid, err := u.tokenSrv.Verify(ctx, opts.RefreshTkn)
	if err != nil {
		return nil, rescode.Failed(err)
	}
	if !isValid {
		return nil, auth.InvalidOrExpiredToken(errors.New("invalid or expired refresh token"))
	}
	session, notFound, err := u.sessionRepo.FindByIds(ctx, trc, claims.Id, state.GetDeviceId(ctx))
	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, auth.InvalidRefreshToken(errors.New("invalid refresh with access token and ip"))
	}
	if !session.IsRefreshValid(opts.AccessTkn, opts.RefreshTkn, opts.IpAddr) {
		return nil, auth.InvalidRefreshToken(errors.New("invalid refresh with access token and ip"))
	}
	return claims, nil
}
