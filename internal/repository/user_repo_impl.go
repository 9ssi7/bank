package repository

import (
	"context"
	"database/sql"

	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/google/uuid"
)

type UserSqlRepo struct {
	syncRepo
	txnSqlRepo
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserSqlRepo {
	return &UserSqlRepo{
		db:         db,
		txnSqlRepo: newTxnSqlRepo(db),
		syncRepo:   newSyncRepo(),
	}
}

func (r *UserSqlRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res.Next() {
		res.Scan(&u.Id, &u.Name, &u.Email, &u.IsActive, &u.TempToken, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	}
	return &u, nil
}

func (r *UserSqlRepo) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	var u user.User
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM users WHERE phone = $1", phone)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res.Next() {
		res.Scan(&u.Id, &u.Name, &u.Email, &u.IsActive, &u.TempToken, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	}
	return &u, nil
}

func (r *UserSqlRepo) FindById(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var u user.User
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res.Next() {
		res.Scan(&u.Id, &u.Name, &u.Email, &u.IsActive, &u.TempToken, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	}
	return &u, nil
}

func (r *UserSqlRepo) FindByToken(ctx context.Context, token string) (*user.User, error) {
	var u user.User
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT * FROM users WHERE temp_token = $1", token)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res.Next() {
		res.Scan(&u.Id, &u.Name, &u.Email, &u.IsActive, &u.TempToken, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	}
	return &u, nil
}

func (r *UserSqlRepo) IsExistsByEmail(ctx context.Context, email string) (bool, error) {
	var total int64
	res, err := r.adapter.GetCurrent().QueryContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email)
	if err != nil {
		return false, err
	}
	defer res.Close()
	if res.Next() {
		res.Scan(&total)
	}
	return total > 0, nil
}

func (r *UserSqlRepo) Save(ctx context.Context, user *user.User) error {
	r.syncRepo.Lock()
	defer r.syncRepo.Unlock()
	q := "INSERT INTO users (id, name, email, is_active, temp_token, verified_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	if user.Id == uuid.Nil {
		user.Id = uuid.New()
		q = "UPDATE users SET name = $2, email = $3, is_active = $4, temp_token = $5, verified_at = $6, created_at = $7, updated_at = $8 WHERE id = $1"
	}
	_, err := r.adapter.GetCurrent().ExecContext(ctx, q, user.Id, user.Name, user.Email, user.IsActive, user.TempToken, user.VerifiedAt, user.CreatedAt, user.UpdatedAt)
	return err
}
