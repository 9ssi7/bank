package accountusecase

import (
	"context"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type AccountSuspend struct {
	UserId    uuid.UUID `json:"user_id" validate:"-"`
	AccountId uuid.UUID `json:"account_id"  params:"account_id" validate:"required,uuid"`
}

type AccountSuspendRes struct{}

type AccountSuspendUseCase usecase.Handler[AccountSuspend, *AccountSuspendRes]

func NewAccountSuspendUseCase(v validation.Service, accountRepo account.Repo) AccountSuspendUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req AccountSuspend) (*AccountSuspendRes, error) {
		ctx = usecase.Push(ctx, tracer, "AccountSuspend")
		if err := v.ValidateStruct(ctx, req); err != nil {
			return nil, err
		}
		account, err := accountRepo.FindByUserIdAndId(ctx, req.UserId, req.AccountId)
		if err != nil {
			return nil, err
		}
		account.Suspend()
		if err := accountRepo.Save(ctx, account); err != nil {
			return nil, err
		}
		return &AccountSuspendRes{}, nil
	}
}
