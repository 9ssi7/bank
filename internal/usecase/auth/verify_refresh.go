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

type VerifyRefresh struct {
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
	IpAddr       string `json:"-"`
}

type VerifyRefreshRes struct {
	User *token.UserClaim
}

type VerifyRefreshUseCase usecase.Handler[VerifyRefresh, *VerifyRefreshRes]

func NewVerifyRefreshUseCase(tokenSrv token.Srv, sessionRepo auth.SessionRepo) VerifyRefreshUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req VerifyRefresh) (*VerifyRefreshRes, error) {
		ctx = usecase.Push(ctx, tracer, "VerifyRefresh")
		claims, err := tokenSrv.Parse(req.RefreshToken)
		if err != nil {
			return nil, rescode.Failed(err)
		}
		isValid, err := tokenSrv.Verify(req.RefreshToken)
		if err != nil {
			return nil, rescode.Failed(err)
		}
		if !isValid {
			return nil, auth.InvalidOrExpiredToken(errors.New("invalid or expired refresh token"))
		}
		session, notFound, err := sessionRepo.FindByIds(ctx, claims.Id, state.GetDeviceId(ctx))
		if err != nil {
			return nil, err
		}
		if notFound {
			return nil, auth.InvalidRefreshToken(errors.New("invalid refresh with access token and ip"))
		}
		if !session.IsRefreshValid(req.AccessToken, req.RefreshToken, req.IpAddr) {
			return nil, auth.InvalidRefreshToken(errors.New("invalid refresh with access token and ip"))
		}
		return &VerifyRefreshRes{
			User: claims,
		}, nil
	}
}
