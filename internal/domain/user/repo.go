package user

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type Repo interface {
	FindByEmail(ctx context.Context, t trace.Tracer, opts FindByEmailOpts) (*User, error)
	FindById(ctx context.Context, t trace.Tracer, opts FindByIdOpts) (*User, error)
	FindByToken(ctx context.Context, t trace.Tracer, opts FindByTokenOpts) (*User, error)
	IsExistsByEmail(ctx context.Context, t trace.Tracer, opts IsExistsByEmailOpts) (bool, error)
	Save(ctx context.Context, t trace.Tracer, opts SaveOpts) error
}

type FindByEmailOpts struct {
	Email string `example:"john@doe.com"`
}

type FindByIdOpts struct {
	ID uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
}

type FindByTokenOpts struct {
	Token string
}

type IsExistsByEmailOpts struct {
	Email string `example:"john@doe.com"`
}

type SaveOpts struct {
	User *User `example:"{}"`
}
