package authusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/state"
	"github.com/9ssi7/bank/pkg/usecase"
	"go.opentelemetry.io/otel/trace"
)

type LoginCheck struct {
	VerifyToken string `json:"-"`
}

type LoginCheckRes struct{}

type LoginCheckUseCase usecase.Handler[LoginCheck, *LoginCheckRes]

func NewLoginCheckUseCase(verifyRepo auth.VerifyRepo) LoginCheckUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req LoginCheck) (*LoginCheckRes, error) {
		ctx = usecase.Push(ctx, tracer, "LoginCheck")
		exists, err := verifyRepo.IsExists(ctx, req.VerifyToken, state.GetDeviceId(ctx))
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, rescode.NotFound(errors.New("verify token not exists"))
		}
		return &LoginCheckRes{}, nil
	}
}
