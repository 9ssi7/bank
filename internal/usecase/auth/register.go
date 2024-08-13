package authusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"go.opentelemetry.io/otel/trace"
)

type Register struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type RegisterRes struct{}

type RegisterUseCase usecase.Handler[Register, *RegisterRes]

func NewRegisterUseCase(v validation.Service, userRepo user.Repo, eventer eventer.Srv) RegisterUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req Register) (*RegisterRes, error) {
		ctx = usecase.Push(ctx, tracer, "Register")
		err := v.ValidateStruct(ctx, req)
		if err != nil {
			return nil, err
		}
		exists, err := userRepo.IsExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, user.EmailAlreadyExists(errors.New("email already exists"))
		}
		u := user.New(req.Name, req.Email)
		err = userRepo.Save(ctx, u)
		if err != nil {
			return nil, err
		}
		eventer.Publish(ctx, user.SubjectCreated, &user.EventCreated{
			Name:      u.Name,
			Email:     u.Email,
			TempToken: *u.TempToken,
		})
		return &RegisterRes{}, nil
	}
}
