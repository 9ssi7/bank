package repository_test

import (
	"context"
	"testing"
)

func TestAllRepositories(t *testing.T) {
	ctx := context.Background()
	db, cancel := createSqlTesting(t)
	defer cancel()

	t.Run("AccountRepo", func(t *testing.T) {
		testAccountRepo(ctx, db, t)
	})

	t.Run("TransactionRepo", func(t *testing.T) {
		testTransactionRepo(ctx, db, t)
	})
}
