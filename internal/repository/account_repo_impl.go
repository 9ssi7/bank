package repository

import (
	"context"
	"database/sql"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/list"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
)

type AccountSqlRepo struct {
	syncRepo
	txnSqlRepo
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) *AccountSqlRepo {
	return &AccountSqlRepo{
		db:         db,
		txnSqlRepo: newTxnSqlRepo(db),
		syncRepo:   newSyncRepo(),
	}
}

func (r *AccountSqlRepo) Save(ctx context.Context, t trace.Tracer, opts account.SaveOpts) error {
	ctx, span := t.Start(ctx, "AccountSqlRepo.Save")
	defer span.End()
	r.syncRepo.Lock()
	defer r.syncRepo.Unlock()
	q := "INSERT INTO accounts (id, user_id, iban, owner, balance, currency) VALUES ($1, $2, $3, $4, $5, $6)"
	if opts.Acount.ID == uuid.Nil {
		opts.Acount.ID = uuid.New()
		q = "UPDATE accounts SET user_id = $2, iban = $3, owner = $4, balance = $5, currency = $6 WHERE id = $1"
	}
	_, err := r.adapter.GetCurrent().ExecContext(ctx, q, opts.Acount.ID, opts.Acount.UserId, opts.Acount.Iban, opts.Acount.Owner, opts.Acount.Balance, opts.Acount.Currency)
	return err
}

func (r *AccountSqlRepo) ListByUserId(ctx context.Context, t trace.Tracer, opts account.ListByUserIdOpts) (*list.PagiResponse[*account.Account], error) {
	ctx, span := t.Start(ctx, "AccountSqlRepo.ListByUserId")
	defer span.End()
	var total int64
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT COUNT(*) FROM accounts WHERE user_id = $1", opts.UserId)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&total)
	}
	res.Close()
	accounts := make([]*account.Account, 0)
	res, err = r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE user_id = $1 LIMIT $2 OFFSET $3", opts.UserId, *opts.Pagi.Limit, opts.Pagi.Offset())
	if err != nil {
		return nil, err
	}
	for res.Next() {
		var a account.Account
		res.Scan(&a.ID, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
		accounts = append(accounts, &a)
	}
	res.Close()
	return &list.PagiResponse[*account.Account]{
		List:          accounts,
		Total:         total,
		Limit:         *opts.Pagi.Limit,
		Page:          *opts.Pagi.Page,
		FilteredTotal: total,
		TotalPage:     opts.Pagi.TotalPage(total),
	}, nil
}

func (r *AccountSqlRepo) FindByIbanAndOwner(ctx context.Context, t trace.Tracer, opts account.FindByIbanAndOwnerOpts) (*account.Account, error) {
	ctx, span := t.Start(ctx, "AccountSqlRepo.FindByIbanAndOwner")
	defer span.End()
	var a account.Account
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE iban = $1 AND owner = $2", opts.Iban, opts.Owner)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&a.ID, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
	}
	res.Close()
	return &a, nil
}

func (r *AccountSqlRepo) FindByUserIdAndId(ctx context.Context, t trace.Tracer, opts account.FindByUserIdAndIdOpts) (*account.Account, error) {
	ctx, span := t.Start(ctx, "AccountSqlRepo.FindByUserIdAndId")
	defer span.End()
	var a account.Account
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE user_id = $1 AND id = $2", opts.UserId, opts.ID)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&a.ID, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
	}
	res.Close()
	return &a, nil
}

func (r *AccountSqlRepo) FindById(ctx context.Context, t trace.Tracer, opts account.FindByIdOpts) (*account.Account, error) {
	ctx, span := t.Start(ctx, "AccountSqlRepo.FindByUserIdAndId")
	defer span.End()
	var a account.Account
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE id = $1", opts.ID)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&a.ID, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
	}
	res.Close()
	return &a, nil
}
