package repository

import (
	"context"
	"database/sql"

	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
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

func (r *UserSqlRepo) FindByEmail(ctx context.Context, trc trace.Tracer, email string) (*user.User, error) {
	ctx, span := trc.Start(ctx, "UserSqlRepo.FindByEmail")
	defer span.End()
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

func (r *UserSqlRepo) FindByPhone(ctx context.Context, trc trace.Tracer, phone string) (*user.User, error) {
	ctx, span := trc.Start(ctx, "UserSqlRepo.FindByPhone")
	defer span.End()
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

func (r *UserSqlRepo) FindById(ctx context.Context, trc trace.Tracer, id uuid.UUID) (*user.User, error) {
	ctx, span := trc.Start(ctx, "UserSqlRepo.FindById")
	defer span.End()
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

func (r *UserSqlRepo) FindByToken(ctx context.Context, trc trace.Tracer, token string) (*user.User, error) {
	ctx, span := trc.Start(ctx, "UserSqlRepo.FindByToken")
	defer span.End()
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

func (r *UserSqlRepo) IsExistsByEmail(ctx context.Context, trc trace.Tracer, email string) (bool, error) {
	ctx, span := trc.Start(ctx, "UserSqlRepo.IsExistsByEmail")
	defer span.End()
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

func (r *UserSqlRepo) Save(ctx context.Context, trc trace.Tracer, user *user.User) error {
	ctx, span := trc.Start(ctx, "UserSqlRepo.Save")
	defer span.End()
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
