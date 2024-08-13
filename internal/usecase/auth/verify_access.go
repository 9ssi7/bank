package authusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/state"
	"github.com/9ssi7/bank/pkg/token"
	"github.com/9ssi7/bank/pkg/usecase"
	"go.opentelemetry.io/otel/trace"
)

type VerifyAccess struct {
	AccessToken  string `json:"-"`
	IpAddr       string `json:"-"`
	IsUnverified bool   `json:"-"`
}

type VerifyAccessRes struct {
	User *token.UserClaim
}

type VerifyAccessUseCase usecase.Handler[VerifyAccess, *VerifyAccessRes]

func NewVerifyAccessUseCase(tokenSrv token.Srv, sessionRepo auth.SessionRepo) VerifyAccessUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req VerifyAccess) (*VerifyAccessRes, error) {
		ctx = usecase.Push(ctx, tracer, "VerifyAccess")
		var claims *token.UserClaim
		var err error
		if req.IsUnverified {
			claims, err = tokenSrv.Parse(req.AccessToken)
		} else {
			claims, err = tokenSrv.VerifyAndParse(req.AccessToken)
		}
		if err != nil {
			return nil, rescode.Failed(err)
		}
		session, notExists, err := sessionRepo.FindByIds(ctx, claims.Id, state.GetDeviceId(ctx))
		if err != nil {
			return nil, err
		}
		if notExists {
			return nil, auth.InvalidAccess(errors.New("invalid access with token and ip"))
		}
		if !session.IsAccessValid(req.AccessToken, req.IpAddr) {
			return nil, auth.InvalidAccess(errors.New("invalid access with token and ip"))
		}
		return &VerifyAccessRes{
			User: claims,
		}, nil
	}
}
