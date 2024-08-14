package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/repository"
	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/ptr"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/trace"
)

func testTransactionRepo(ctx context.Context, db *sql.DB, trc trace.Tracer, t *testing.T) {
	repo := repository.NewTransactionRepo(db)

	t.Run("Create", func(t *testing.T) {
		accountId := uuid.New()
		amount := decimal.NewFromFloat(100)
		tx := account.NewTransaction(accountId, accountId, amount, "test", account.TransactionKindDeposit)
		err := repo.Save(ctx, trc, tx)
		if err != nil {
			t.Fatalf("Could not save transaction: %s", err)
		}
		if tx.Id == uuid.Nil {
			t.Fatalf("Transaction id is empty")
		}
	})

	t.Run("Filter", func(t *testing.T) {
		accountId := uuid.New()
		amount := decimal.NewFromFloat(100)
		tx := account.NewTransaction(accountId, accountId, amount, "test", account.TransactionKindDeposit)
		err := repo.Save(ctx, trc, tx)
		if err != nil {
			t.Fatalf("Could not save transaction: %s", err)
		}
		pagi := &list.PagiRequest{Limit: ptr.Int(10), Page: ptr.Int(1)}
		filters := &account.TransactionFilters{}
		_, err = repo.Filter(ctx, trc, accountId, pagi, filters)
		if err != nil {
			t.Fatalf("Could not filter transaction: %s", err)
		}
	})
}
