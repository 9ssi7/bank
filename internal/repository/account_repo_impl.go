package repository

import (
	"context"
	"database/sql"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/list"

	"github.com/google/uuid"
)

type accountRepo struct {
	syncRepo
	txnSqlRepo
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) account.Repo {
	return &accountRepo{
		db:         db,
		txnSqlRepo: newTxnSqlRepo(db),
	}
}

func (r *accountRepo) Save(ctx context.Context, account *account.Account) error {
	r.syncRepo.Lock()
	defer r.syncRepo.Unlock()
	q := "INSERT INTO accounts (id, user_id, iban, owner, balance, currency) VALUES ($1, $2, $3, $4, $5, $6)"
	if account.Id == uuid.Nil {
		account.Id = uuid.New()
		q = "UPDATE accounts SET user_id = $2, iban = $3, owner = $4, balance = $5, currency = $6 WHERE id = $1"
	}
	_, err := r.adapter.GetCurrent().ExecContext(ctx, q, account.Id, account.UserId, account.Iban, account.Owner, account.Balance, account.Currency)
	return err
}

func (r *accountRepo) ListByUserId(ctx context.Context, userId uuid.UUID, pagi *list.PagiRequest) (*list.PagiResponse[*account.Account], error) {
	var total int64
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT COUNT(*) FROM accounts WHERE user_id = $1", userId)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&total)
	}
	res.Close()
	accounts := make([]*account.Account, 0)
	res, err = r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE user_id = $1 LIMIT $2 OFFSET $3", userId, *pagi.Limit, pagi.Offset())
	if err != nil {
		return nil, err
	}
	for res.Next() {
		var a account.Account
		res.Scan(&a.Id, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
		accounts = append(accounts, &a)
	}
	res.Close()
	return &list.PagiResponse[*account.Account]{
		List:          accounts,
		Total:         total,
		Limit:         *pagi.Limit,
		Page:          *pagi.Page,
		FilteredTotal: total,
		TotalPage:     pagi.TotalPage(total),
	}, nil
}

func (r *accountRepo) FindByIbanAndOwner(ctx context.Context, iban string, owner string) (*account.Account, error) {
	var a account.Account
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE iban = $1 AND owner = $2", iban, owner)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&a.Id, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
	}
	res.Close()
	return &a, nil
}

func (r *accountRepo) FindByUserIdAndId(ctx context.Context, userId uuid.UUID, id uuid.UUID) (*account.Account, error) {
	var a account.Account
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE user_id = $1 AND id = $2", userId, id)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&a.Id, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
	}
	res.Close()
	return &a, nil
}

func (r *accountRepo) FindById(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	var a account.Account
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM accounts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	if res.Next() {
		res.Scan(&a.Id, &a.UserId, &a.Iban, &a.Owner, &a.Balance, &a.Currency)
	}
	res.Close()
	return &a, nil
}
