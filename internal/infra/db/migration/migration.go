package migration

import (
	"context"
	"database/sql"
)

type migration func(context.Context, *sql.DB) error

func runner(ctx context.Context, db *sql.DB, migrations ...migration) error {
	for _, m := range migrations {
		if err := m(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

func Run(ctx context.Context, db *sql.DB) error {
	return runner(ctx, db, userModelMigration, accountModelMigration, transactionModelMigration)
}

func userModelMigration(ctx context.Context, db *sql.DB) error {
	q := `CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT FALSE,
		temp_token VARCHAR(255) NOT NULL,
		verified_at TIMESTAMP NULL DEFAULT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL DEFAULT NULL
	)`
	_, err := db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)`
	_, err = db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE INDEX IF NOT EXISTS idx_users_temp_token ON users (temp_token)`
	_, err = db.ExecContext(ctx, q)
	return err
}

func accountModelMigration(ctx context.Context, db *sql.DB) error {
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
		return err
	}
	q = `CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts (user_id)`
	_, err = db.ExecContext(ctx, q)
	return err
}

func transactionModelMigration(ctx context.Context, db *sql.DB) error {
	q := `CREATE TABLE IF NOT EXISTS transactions (
		id UUID PRIMARY KEY,
		sender_id UUID NOT NULL,
		receiver_id UUID NOT NULL,
		amount DECIMAL(10, 2) NOT NULL,
		description TEXT NOT NULL,
		kind VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE INDEX IF NOT EXISTS idx_transactions_sender_id ON transactions (sender_id, receiver_id)`
	_, err = db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE INDEX IF NOT EXISTS idx_transactions_receiver_id ON transactions (kind)`
	_, err = db.ExecContext(ctx, q)
	return err
}
