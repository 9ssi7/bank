package accountusecase

import (
	"context"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type AccountLock struct {
	UserId    uuid.UUID `json:"user_id" validate:"-"`
	AccountId uuid.UUID `json:"account_id"  params:"account_id" validate:"required,uuid"`
}

type AccountLockRes struct{}

type AccountLockUseCase usecase.Handler[AccountLock, *AccountLockRes]

func NewAccountLockUseCase(v validation.Service, accountRepo account.Repo) AccountLockUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req AccountLock) (*AccountLockRes, error) {
		ctx = usecase.Push(ctx, tracer, "AccountLock")
		if err := v.ValidateStruct(ctx, req); err != nil {
			return nil, err
		}
		account, err := accountRepo.FindByUserIdAndId(ctx, req.UserId, req.AccountId)
		if err != nil {
			return nil, err
		}
		account.Lock()
		if err := accountRepo.Save(ctx, account); err != nil {
			return nil, err
		}
		return &AccountLockRes{}, nil
	}
}
