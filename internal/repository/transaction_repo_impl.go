package repository

import (
	"context"
	"database/sql"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/query"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type TransactionSqlRepo struct {
	syncRepo
	txnSqlRepo
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) *TransactionSqlRepo {
	return &TransactionSqlRepo{
		db:         db,
		txnSqlRepo: newTxnSqlRepo(db),
		syncRepo:   newSyncRepo(),
	}
}

func (r *TransactionSqlRepo) Save(ctx context.Context, trc trace.Tracer, opts account.TransactionSaveOpts) error {
	ctx, span := trc.Start(ctx, "TransactionSqlRepo.Save")
	defer span.End()
	r.syncRepo.Lock()
	defer r.syncRepo.Unlock()
	q := "INSERT INTO transactions (id, sender_id, receiver_id, amount, description, kind, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	if opts.Transaction.ID == uuid.Nil {
		opts.Transaction.ID = uuid.New()
		q = "UPDATE transactions SET sender_id = $2, receiver_id = $3, amount = $4, description = $5, kind = $6, created_at = $7 WHERE id = $1"
	}
	_, err := r.adapter.GetCurrent().ExecContext(ctx, q, opts.Transaction.ID, opts.Transaction.SenderId, opts.Transaction.ReceiverId, opts.Transaction.Amount, opts.Transaction.Description, opts.Transaction.Kind, opts.Transaction.CreatedAt)
	return err
}

func (r *TransactionSqlRepo) Filter(ctx context.Context, trc trace.Tracer, opts account.TransactionFilterOpts) (*list.PagiResponse[*account.Transaction], error) {
	ctx, span := trc.Start(ctx, "TransactionSqlRepo.Filter")
	defer span.End()
	var total int64
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT COUNT(*) FROM transactions WHERE sender_id = $1 OR receiver_id = $1", opts.AccountId)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&total)
	}
	res.Close()
	transactions := make([]*account.Transaction, 0)
	conds, vals := query.Build([]query.Conds{
		{
			Key:    "sender_id = ? OR receiver_id = ?",
			Values: query.V{opts.AccountId, opts.AccountId},
			Skip:   false,
		},
		{
			Key:    "kind = ?",
			Values: query.V{opts.Filters.Kind},
			Skip:   opts.Filters.Kind == "",
		},
		{
			Key:    "created_at <= ?",
			Values: query.V{opts.Filters.EndDate},
			Skip:   opts.Filters.EndDate == "",
		},
		{
			Key:    "created_at >= ?",
			Values: query.V{opts.Filters.StartDate},
			Skip:   opts.Filters.StartDate == "",
		},
		{
			Key:    "LIMIT ? OFFSET ?",
			Values: query.V{*opts.Pagi.Limit, opts.Pagi.Offset()},
			Skip:   false,
		},
	})
	q := "SELECT * FROM transactions WHERE " + conds
	res, err = r.adapter.GetCurrent().QueryContext(ctx, q, vals...)
	if err != nil {
		return nil, err
	}
	for res.Next() {
		var t account.Transaction
		res.Scan(&t.ID, &t.SenderId, &t.ReceiverId, &t.Amount, &t.Description, &t.Kind, &t.CreatedAt)
		transactions = append(transactions, &t)
	}
	res.Close()
	return &list.PagiResponse[*account.Transaction]{
		List:          transactions,
		Total:         total,
		Limit:         *opts.Pagi.Limit,
		Page:          *opts.Pagi.Page,
		FilteredTotal: total,
		TotalPage:     opts.Pagi.TotalPage(total),
	}, nil
}
