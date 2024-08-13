package accountusecase

import (
	"context"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type AccountActivate struct {
	UserId    uuid.UUID `json:"user_id" validate:"-"`
	AccountId uuid.UUID `json:"account_id"  params:"account_id" validate:"required,uuid"`
}

type AccountActivateRes struct{}

type AccountActivateUseCase usecase.Handler[AccountActivate, *AccountActivateRes]

func NewAccountActivateUseCase(v validation.Service, accountRepo account.Repo) AccountActivateUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req AccountActivate) (*AccountActivateRes, error) {
		ctx = usecase.Push(ctx, tracer, "AccountActivate")
		if err := v.ValidateStruct(ctx, req); err != nil {
			return nil, err
		}
		account, err := accountRepo.FindByUserIdAndId(ctx, req.UserId, req.AccountId)
		if err != nil {
			return nil, err
		}
		account.Activate()
		if err := accountRepo.Save(ctx, account); err != nil {
			return nil, err
		}
		return &AccountActivateRes{}, nil
	}
}
