package repository_test

import (
	"context"
	"testing"

	"github.com/9ssi7/bank/test/tracertest"
)

func TestAllRepositories(t *testing.T) {
	ctx := context.Background()
	db, cancel := createSqlTesting(t)
	defer cancel()

	tracer := tracertest.CreateTracerTesting()

	t.Run("AccountRepo", func(t *testing.T) {
		testAccountRepo(ctx, db, tracer, t)
	})

	t.Run("TransactionRepo", func(t *testing.T) {
		testTransactionRepo(ctx, db, tracer, t)
	})
}
