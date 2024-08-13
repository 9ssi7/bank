package account

import (
	"context"

	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/txadapter"
	"github.com/google/uuid"
)

type Repo interface {
	txadapter.Repo

	Save(ctx context.Context, account *Account) error
	ListByUserId(ctx context.Context, userId uuid.UUID, pagi *list.PagiRequest) (*list.PagiResponse[*Account], error)
	FindByIbanAndOwner(ctx context.Context, iban string, owner string) (*Account, error)
	FindByUserIdAndId(ctx context.Context, userId uuid.UUID, id uuid.UUID) (*Account, error)
	FindById(ctx context.Context, id uuid.UUID) (*Account, error)
}

type TransactionRepo interface {
	txadapter.Repo

	Save(ctx context.Context, transaction *Transaction) error
	Filter(ctx context.Context, accountId uuid.UUID, pagi *list.PagiRequest, filters *TransactionFilters) (*list.PagiResponse[*Transaction], error)
}
