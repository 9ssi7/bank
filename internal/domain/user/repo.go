package user

import (
	"context"

	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/txadapter"
	"github.com/google/uuid"
)

type Repo interface {
	txadapter.Repo

	Save(ctx context.Context, user *User) error
	IsExistsByEmail(ctx context.Context, email string) (bool, error)
	FindByToken(ctx context.Context, token string) (*User, error)
	FindById(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Filter(ctx context.Context, req *list.PagiRequest) (*list.PagiResponse[*User], error)
}
