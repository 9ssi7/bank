package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/repository"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

func testAccountRepo(ctx context.Context, db *sql.DB, trc trace.Tracer, t *testing.T) {
	repo := repository.NewAccountRepo(db)

	t.Run("Create", func(t *testing.T) {
		userId := uuid.New()
		acc := account.New(userId, "test", "test 0", "TRY")
		err := repo.Save(ctx, trc, acc)
		if err != nil {
			t.Fatalf("Could not save account: %s", err)
		}
		if acc.Id == uuid.Nil {
			t.Fatalf("Account id is empty")
		}
	})

	t.Run("Update", func(t *testing.T) {
		userId := uuid.New()
		acc := account.New(userId, "test", "test 0", "TRY")
		err := repo.Save(ctx, trc, acc)
		if err != nil {
			t.Fatalf("Could not save account: %s", err)
		}
		acc.Owner = "test 1"
		err = repo.Save(ctx, trc, acc)
		if err != nil {
			t.Fatalf("Could not update account: %s", err)
		}
		if acc.Owner != "test 1" {
			t.Fatalf("Account owner is not updated")
		}
	})
}
