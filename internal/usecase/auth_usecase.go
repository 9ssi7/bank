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
	TokenSrv    TokenSrv
	EventSrv    EventSrv
	VerifyRepo  auth.VerifyRepo
	UserRepo    user.Repo
	SessionRepo auth.SessionRepo
}

type AuthLoginVerifyCheckOpts struct {
	VerifyToken string
}

func (u *AuthUseCase) LoginVerifyCheck(ctx context.Context, trc trace.Tracer, opts AuthLoginVerifyCheckOpts) error {
	ctx, span := trc.Start(ctx, "AuthUseCase.VerifyCheck")
	defer span.End()
	exists, err := u.VerifyRepo.IsExists(ctx, trc, auth.VerifyIsExistsOpts{
		Token:    opts.VerifyToken,
		DeviceId: state.GetDeviceId(ctx),
	})
	if err != nil {
		return err
	}
	if !exists {
		return rescode.NotFound(errors.New("verify token not exists"))
	}
	return nil
}

type AuthLoginStartOpts struct {
	Email  string
	Device agent.Device
}

func (u *AuthUseCase) LoginStart(ctx context.Context, trc trace.Tracer, opts AuthLoginStartOpts) (*string, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.LoginStart")
	defer span.End()
	usr, err := u.UserRepo.FindByEmail(ctx, trc, user.FindByEmailOpts{Email: opts.Email})
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
		UserId:   usr.ID,
		DeviceId: state.GetDeviceId(ctx),
		Locale:   state.GetLocale(ctx),
	})
	if err := u.VerifyRepo.Save(ctx, trc, auth.VerifySaveOpts{Token: verifyToken, Verify: verify}); err != nil {
		return nil, err
	}
	err = u.EventSrv.Publish(ctx, auth.SubjectLoginStarted, &auth.EventLoginStarted{
		Email:  usr.Email,
		Code:   verify.Code,
		Device: opts.Device,
	})
	if err != nil {
		return nil, err
	}
	return &verifyToken, nil
}

type AuthLoginVerifyOpts struct {
	Code        string
	VerifyToken string
	Device      agent.Device
}

func (u *AuthUseCase) LoginVerify(ctx context.Context, trc trace.Tracer, opts AuthLoginVerifyOpts) (*string, *string, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.LoginVerify")
	defer span.End()
	verify, err := u.VerifyRepo.Find(ctx, trc, auth.VerifyFindOpts{Token: opts.VerifyToken, DeviceId: state.GetDeviceId(ctx)})
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
		err = u.VerifyRepo.Save(ctx, trc, auth.VerifySaveOpts{Token: opts.VerifyToken, Verify: verify})
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, auth.VerificationInvalid(errors.New("verification invalid"))
	}
	err = u.VerifyRepo.Delete(ctx, trc, auth.VerifyDeleteOpts{Token: opts.VerifyToken, DeviceId: state.GetDeviceId(ctx)})
	if err != nil {
		return nil, nil, err
	}
	user, err := u.UserRepo.FindById(ctx, trc, user.FindByIdOpts{ID: verify.UserId})
	if err != nil {
		return nil, nil, err
	}
	claims := token.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
	accessToken, err := u.TokenSrv.GenerateAccessToken(ctx, claims)
	if err != nil {
		return nil, nil, rescode.Failed(err)
	}
	refreshToken, err := u.TokenSrv.GenerateRefreshToken(ctx, claims)
	if err != nil {
		return nil, nil, rescode.Failed(err)
	}
	ses := auth.NewSession(auth.SessionConfig{
		Device:       opts.Device,
		DeviceId:     state.GetDeviceId(ctx),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
	if err = u.SessionRepo.Save(ctx, trc, auth.SessionSaveOpts{UserId: user.ID, Session: ses}); err != nil {
		return nil, nil, err
	}
	return &accessToken, &refreshToken, nil
}

type AuthRefreshTokenOpts struct {
	UserId     uuid.UUID
	AccessTkn  string
	RefreshTkn string
	IpAddr     string
}

func (u *AuthUseCase) RefreshToken(ctx context.Context, trc trace.Tracer, opts AuthRefreshTokenOpts) (*string, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.RefreshToken")
	defer span.End()
	session, notFound, err := u.SessionRepo.Find(ctx, trc, auth.SessionFindOpts{UserId: opts.UserId, DeviceId: state.GetDeviceId(ctx)})
	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, auth.InvalidRefreshOrAccessTokens(errors.New("invalid refresh with access token and ip"))
	}
	if !session.IsRefreshValid(opts.AccessTkn, opts.RefreshTkn, opts.IpAddr) {
		return nil, auth.InvalidRefreshOrAccessTokens(errors.New("invalid refresh with access token and ip"))
	}
	user, err := u.UserRepo.FindById(ctx, trc, user.FindByIdOpts{ID: opts.UserId})
	if err != nil {
		return nil, err
	}
	claims := token.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
	accessToken, err := u.TokenSrv.GenerateAccessToken(ctx, claims)
	if err != nil {
		return nil, rescode.Failed(err)
	}
	session.Refresh(accessToken)
	if err := u.SessionRepo.Save(ctx, trc, auth.SessionSaveOpts{UserId: user.ID, Session: session}); err != nil {
		return nil, err
	}
	return &accessToken, nil
}

type AuthRegisterOpts struct {
	Name  string
	Email string
}

func (u *AuthUseCase) Register(ctx context.Context, trc trace.Tracer, opts AuthRegisterOpts) error {
	ctx, span := trc.Start(ctx, "AuthUseCase.Register")
	defer span.End()
	exists, err := u.UserRepo.IsExistsByEmail(ctx, trc, user.IsExistsByEmailOpts{Email: opts.Email})
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
	err = u.UserRepo.Save(ctx, trc, user.SaveOpts{User: usr})
	if err != nil {
		return err
	}
	err = u.EventSrv.Publish(ctx, user.SubjectCreated, &user.EventCreated{
		Name:      opts.Name,
		Email:     opts.Email,
		TempToken: *usr.TempToken,
	})
	if err != nil {
		return err
	}
	return nil
}

type AuthRegistrationVerifyOpts struct {
	Token string
}

func (u *AuthUseCase) RegistrationVerify(ctx context.Context, trc trace.Tracer, opts AuthRegistrationVerifyOpts) error {
	ctx, span := trc.Start(ctx, "AuthUseCase.RegistrationVerify")
	defer span.End()
	usr, err := u.UserRepo.FindByToken(ctx, trc, user.FindByTokenOpts{Token: opts.Token})
	if err != nil {
		return err
	}
	usr.Verify()
	return u.UserRepo.Save(ctx, trc, user.SaveOpts{User: usr})
}

type AuthVerifyAccessOpts struct {
	AccessTkn  string
	IpAddr     string
	SkipVerify bool
}

func (u *AuthUseCase) VerifyAccess(ctx context.Context, trc trace.Tracer, opts AuthVerifyAccessOpts) (*token.UserClaim, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.VerifyAccess")
	defer span.End()
	var claims *token.UserClaim
	var err error
	if opts.SkipVerify {
		claims, err = u.TokenSrv.Parse(ctx, opts.AccessTkn)
	} else {
		claims, err = u.TokenSrv.VerifyAndParse(ctx, opts.AccessTkn)
	}
	if err != nil {
		return nil, rescode.Failed(err)
	}
	session, notExists, err := u.SessionRepo.Find(ctx, trc, auth.SessionFindOpts{UserId: claims.User.ID, DeviceId: state.GetDeviceId(ctx)})
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

type AuthVerifyRefreshOpts struct {
	AccessTkn  string
	RefreshTkn string
	IpAddr     string
}

func (u *AuthUseCase) VerifyRefresh(ctx context.Context, trc trace.Tracer, opts *AuthVerifyRefreshOpts) (*token.UserClaim, error) {
	ctx, span := trc.Start(ctx, "AuthUseCase.VerifyRefresh")
	defer span.End()
	claims, err := u.TokenSrv.Parse(ctx, opts.RefreshTkn)
	if err != nil {
		return nil, rescode.Failed(err)
	}
	isValid, err := u.TokenSrv.Verify(ctx, opts.RefreshTkn)
	if err != nil {
		return nil, rescode.Failed(err)
	}
	if !isValid {
		return nil, auth.InvalidOrExpiredToken(errors.New("invalid or expired refresh token"))
	}
	session, notFound, err := u.SessionRepo.Find(ctx, trc, auth.SessionFindOpts{UserId: claims.User.ID, DeviceId: state.GetDeviceId(ctx)})
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
