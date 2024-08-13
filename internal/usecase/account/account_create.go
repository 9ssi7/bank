package accountusecase

import (
	"context"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type AccountCreate struct {
	UserId   uuid.UUID `json:"user_id" validate:"-"`
	Name     string    `json:"name" validate:"required,min=3,max=255"`
	Owner    string    `json:"owner" validate:"required,min=3,max=255"`
	Currency string    `json:"currency" validate:"required,currency"`
}

type AccountCreateUseCase usecase.Handler[AccountCreate, *uuid.UUID]

func NewAccountCreateUseCase(v validation.Service, accountRepo account.Repo) AccountCreateUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req AccountCreate) (*uuid.UUID, error) {
		ctx = usecase.Push(ctx, tracer, "AccountCreate")
		if err := v.ValidateStruct(ctx, req); err != nil {
			return nil, err
		}
		account := account.New(req.UserId, req.Name, req.Owner, req.Currency)
		if err := accountRepo.Save(ctx, account); err != nil {
			return nil, err
		}
		return &account.Id, nil
	}
}
