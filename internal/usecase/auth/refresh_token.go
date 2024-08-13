package authusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/state"
	"github.com/9ssi7/bank/pkg/token"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type RefreshToken struct {
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	IpAddress    string    `json:"-"`
	UserId       uuid.UUID `json:"-"`
}

type RefreshTokenRes struct {
	AccessToken string
}

type RefreshTokenUseCase usecase.Handler[RefreshToken, *RefreshTokenRes]

func NewRefreshTokenUseCase(v validation.Service, tokenSrv token.Srv, sessionRepo auth.SessionRepo, userRepo user.Repo) RefreshTokenUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req RefreshToken) (*RefreshTokenRes, error) {
		ctx = usecase.Push(ctx, tracer, "RefreshToken")
		session, notFound, err := sessionRepo.FindByIds(ctx, req.UserId, state.GetDeviceId(ctx))
		if err != nil {
			return nil, err
		}
		if notFound {
			return nil, auth.InvalidRefreshOrAccessTokens(errors.New("invalid refresh with access token and ip"))
		}
		if !session.IsRefreshValid(req.AccessToken, req.RefreshToken, req.IpAddress) {
			return nil, auth.InvalidRefreshOrAccessTokens(errors.New("invalid refresh with access token and ip"))
		}
		user, err := userRepo.FindById(ctx, req.UserId)
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
		if err != nil {
			return nil, err
		}
		session.Refresh(accessToken)
		if err := sessionRepo.Save(ctx, user.Id, session); err != nil {
			return nil, err
		}
		return &RefreshTokenRes{
			AccessToken: accessToken,
		}, nil
	}
}
