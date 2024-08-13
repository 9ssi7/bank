package authusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/pkg/agent"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/state"
	"github.com/9ssi7/bank/pkg/token"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"go.opentelemetry.io/otel/trace"
)

type LoginVerify struct {
	Code        string        `json:"code" validate:"required,numeric,len=4"`
	VerifyToken string        `json:"-"`
	Device      *agent.Device `json:"-"`
}

type LoginVerifyRes struct {
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
}

type LoginVerifyUseCase usecase.Handler[LoginVerify, *LoginVerifyRes]

func NewLoginVerifyUseCase(v validation.Service, tokenSrv token.Srv, userRepo user.Repo, verifyRepo auth.VerifyRepo, sessionRepo auth.SessionRepo) LoginVerifyUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req LoginVerify) (*LoginVerifyRes, error) {
		ctx = usecase.Push(ctx, tracer, "LoginVerify")
		err := v.ValidateStruct(ctx, req)
		if err != nil {
			return nil, err
		}
		verify, err := verifyRepo.Find(ctx, req.VerifyToken, state.GetDeviceId(ctx))
		if err != nil {
			return nil, err
		}
		if verify.IsExpired() {
			return nil, auth.VerificationExpired(errors.New("verification expired"))
		}
		if verify.IsExceeded() {
			return nil, auth.VerificationExceeded(errors.New("verification exceeded"))
		}
		if req.Code != verify.Code {
			verify.IncTryCount()
			err = verifyRepo.Save(ctx, req.VerifyToken, verify)
			if err != nil {
				return nil, err
			}
			return nil, auth.VerificationInvalid(errors.New("verification invalid"))
		}
		err = verifyRepo.Delete(ctx, req.VerifyToken, state.GetDeviceId(ctx))
		if err != nil {
			return nil, err
		}
		user, err := userRepo.FindById(ctx, verify.UserId)
		if err != nil {
			return nil, err
		}
		claims := token.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		}
		accessToken, err := tokenSrv.GenerateAccessToken(claims)
		if err != nil {
			return nil, rescode.Failed(err)
		}
		refreshToken, err := tokenSrv.GenerateRefreshToken(claims)
		if err != nil {
			return nil, rescode.Failed(err)
		}
		ses := auth.NewSession(*req.Device, state.GetDeviceId(ctx), accessToken, refreshToken)
		if err = sessionRepo.Save(ctx, user.Id, ses); err != nil {
			return nil, err
		}
		return &LoginVerifyRes{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	}
}
