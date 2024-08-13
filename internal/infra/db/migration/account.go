package migration

import (
	"context"
	"database/sql"
	"fmt"
)

func runAccountMigration(ctx context.Context, db *sql.DB) error {
	q := `CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL,
		iban VARCHAR(255) NOT NULL,
		owner VARCHAR(255) NOT NULL,
		balance DECIMAL(10, 2) NOT NULL,
		currency VARCHAR(3) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL DEFAULT NULL
	)`
	_, err := db.ExecContext(ctx, q)
	if err != nil {
		fmt.Println(err)
		return err
	}
	q = `CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts (user_id)`
	_, err = db.ExecContext(ctx, q)
	return err
}
