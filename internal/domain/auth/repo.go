package auth

import (
	"context"

	"github.com/google/uuid"
)

type SessionRepo interface {
	Save(ctx context.Context, userId uuid.UUID, session *Session) error
	FindByIds(ctx context.Context, userId uuid.UUID, deviceId string) (*Session, bool, error)
	FindAllByUserId(ctx context.Context, userId uuid.UUID) ([]*Session, error)
	Destroy(ctx context.Context, userId uuid.UUID, deviceId string) error
}

type VerifyRepo interface {
	Save(ctx context.Context, token string, verify *Verify) error
	IsExists(ctx context.Context, token string, deviceId string) (bool, error)
	Find(ctx context.Context, token string, deviceId string) (*Verify, error)
	Delete(ctx context.Context, token string, deviceId string) error
}
