package authusecase

import (
	"context"

	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"go.opentelemetry.io/otel/trace"
)

type RegistrationVerify struct {
	Token string `json:"token" validate:"required,uuid"`
}

type RegistrationVerifyRes struct{}

type RegistrationVerifyUseCase usecase.Handler[RegistrationVerify, *RegistrationVerifyRes]

func NewRegistrationVerifyUseCase(v validation.Service, userRepo user.Repo) RegistrationVerifyUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req RegistrationVerify) (*RegistrationVerifyRes, error) {
		ctx = usecase.Push(ctx, tracer, "RegistrationVerify")
		u, err := userRepo.FindByToken(ctx, req.Token)
		if err != nil {
			return nil, err
		}
		u.Verify()
		err = userRepo.Save(ctx, u)
		if err != nil {
			return nil, err
		}
		return &RegistrationVerifyRes{}, nil
	}
}
