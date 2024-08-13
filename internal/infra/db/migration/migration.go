package migration

import (
	"context"
	"database/sql"
)

func Run(ctx context.Context, db *sql.DB) error {
	if err := runAccountMigration(ctx, db); err != nil {
		return err
	}
	return nil
}
