package auth

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type VerifyRepo interface {
	Save(ctx context.Context, t trace.Tracer, opts VerifySaveOpts) error
	IsExists(ctx context.Context, t trace.Tracer, opts VerifyIsExistsOpts) (bool, error)
	Find(ctx context.Context, t trace.Tracer, opts VerifyFindOpts) (*Verify, error)
	Delete(ctx context.Context, t trace.Tracer, opts VerifyDeleteOpts) error
}

type SessionRepo interface {
	Save(ctx context.Context, t trace.Tracer, opts SessionSaveOpts) error
	Find(ctx context.Context, t trace.Tracer, opts SessionFindOpts) (*Session, bool, error)
	FindAllByUser(ctx context.Context, t trace.Tracer, opts FindAllByUserOpts) ([]*Session, error)
	Destroy(ctx context.Context, t trace.Tracer, opts SessionDestroyOpts) error
}

type VerifySaveOpts struct {
	Token  string  `example:"token"`
	Verify *Verify `example:"{}"`
}

type VerifyIsExistsOpts struct {
	Token    string `example:"token"`
	DeviceId string `example:"device_id"`
}

type VerifyFindOpts struct {
	Token    string `example:"token"`
	DeviceId string `example:"device_id"`
}

type VerifyDeleteOpts struct {
	Token    string `example:"token"`
	DeviceId string `example:"device_id"`
}

type SessionSaveOpts struct {
	UserId  uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
	Session *Session  `example:"{}"`
}

type SessionFindOpts struct {
	UserId   uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
	DeviceId string    `example:"device_id"`
}

type SessionDestroyOpts struct {
	UserId   uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
	DeviceId string    `example:"device_id"`
}

type FindAllByUserOpts struct {
	UserId uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
}
