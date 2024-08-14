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

func (r *TransactionSqlRepo) Save(ctx context.Context, trc trace.Tracer, transaction *account.Transaction) error {
	ctx, span := trc.Start(ctx, "TransactionSqlRepo.Save")
	defer span.End()
	r.syncRepo.Lock()
	defer r.syncRepo.Unlock()
	q := "INSERT INTO transactions (id, sender_id, receiver_id, amount, description, kind, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	if transaction.Id == uuid.Nil {
		transaction.Id = uuid.New()
		q = "UPDATE transactions SET sender_id = $2, receiver_id = $3, amount = $4, description = $5, kind = $6, created_at = $7 WHERE id = $1"
	}
	_, err := r.adapter.GetCurrent().ExecContext(ctx, q, transaction.Id, transaction.SenderId, transaction.ReceiverId, transaction.Amount, transaction.Description, transaction.Kind, transaction.CreatedAt)
	return err
}

func (r *TransactionSqlRepo) Filter(ctx context.Context, trc trace.Tracer, accountId uuid.UUID, pagi *list.PagiRequest, filters *account.TransactionFilters) (*list.PagiResponse[*account.Transaction], error) {
	ctx, span := trc.Start(ctx, "TransactionSqlRepo.Filter")
	defer span.End()
	var total int64
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT COUNT(*) FROM transactions WHERE sender_id = $1 OR receiver_id = $1", accountId)
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
			Values: query.V{accountId, accountId},
			Skip:   false,
		},
		{
			Key:    "kind = ?",
			Values: query.V{filters.Kind},
			Skip:   filters.Kind == "",
		},
		{
			Key:    "created_at <= ?",
			Values: query.V{filters.EndDate},
			Skip:   filters.EndDate == "",
		},
		{
			Key:    "created_at >= ?",
			Values: query.V{filters.StartDate},
			Skip:   filters.StartDate == "",
		},
		{
			Key:    "LIMIT ? OFFSET ?",
			Values: query.V{*pagi.Limit, pagi.Offset()},
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
		res.Scan(&t.Id, &t.SenderId, &t.ReceiverId, &t.Amount, &t.Description, &t.Kind, &t.CreatedAt)
		transactions = append(transactions, &t)
	}
	res.Close()
	return &list.PagiResponse[*account.Transaction]{
		List:          transactions,
		Total:         total,
		Limit:         *pagi.Limit,
		Page:          *pagi.Page,
		FilteredTotal: total,
		TotalPage:     pagi.TotalPage(total),
	}, nil
}
