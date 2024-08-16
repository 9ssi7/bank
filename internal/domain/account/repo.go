package account

import (
	"context"

	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/txadapter"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type Repo interface {
	txadapter.Repo
	Save(ctx context.Context, t trace.Tracer, opts SaveOpts) error
	ListByUserId(ctx context.Context, t trace.Tracer, opts ListByUserIdOpts) (*list.PagiResponse[*Account], error)
	FindByIbanAndOwner(ctx context.Context, t trace.Tracer, opts FindByIbanAndOwnerOpts) (*Account, error)
	FindByUserIdAndId(ctx context.Context, t trace.Tracer, opts FindByUserIdAndIdOpts) (*Account, error)
	FindById(ctx context.Context, t trace.Tracer, opts FindByIdOpts) (*Account, error)
}

type TransactionRepo interface {
	txadapter.Repo
	Save(ctx context.Context, t trace.Tracer, opts TransactionSaveOpts) error
	Filter(ctx context.Context, t trace.Tracer, opts TransactionFilterOpts) (*list.PagiResponse[*Transaction], error)
}

type SaveOpts struct {
	Acount *Account `example:"{}"`
}

type ListByUserIdOpts struct {
	UserId uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
	Pagi   *list.PagiRequest
}

type FindByIbanAndOwnerOpts struct {
	Iban  string `example:"TR0000000000000000000000"`
	Owner string `example:"John Doe"`
}

type FindByUserIdAndIdOpts struct {
	UserId uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
	ID     uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
}

type FindByIdOpts struct {
	ID uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
}

type TransactionSaveOpts struct {
	Transaction *Transaction `example:"{}"`
}

type TransactionFilterOpts struct {
	AccountId uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
	Pagi      *list.PagiRequest
	Filters   *TransactionFilters
}
