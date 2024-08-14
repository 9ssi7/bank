package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

type VerifyRedisRepo struct {
	syncRepo
	db *redis.Client
}

func NewVerifyRedisRepo(db *redis.Client) *VerifyRedisRepo {
	return &VerifyRedisRepo{
		db: db,
	}
}

func (r *VerifyRedisRepo) Save(ctx context.Context, trc trace.Tracer, token string, verify *auth.Verify) error {
	ctx, span := trc.Start(ctx, "VerifyRedisRepo.Save")
	defer span.End()
	r.syncRepo.Lock()
	defer r.syncRepo.Unlock()
	b, err := json.Marshal(verify)
	if err != nil {
		return rescode.Failed(err)
	}
	if err = r.db.SetEx(ctx, r.calcKey(verify.DeviceId, token), b, 5*time.Minute).Err(); err != nil {
		return rescode.Failed(err)
	}
	return nil
}

func (r *VerifyRedisRepo) IsExists(ctx context.Context, trc trace.Tracer, token string, deviceId string) (bool, error) {
	ctx, span := trc.Start(ctx, "VerifyRedisRepo.IsExists")
	defer span.End()
	res, err := r.db.Get(ctx, r.calcKey(deviceId, token)).Result()
	if err != nil {
		return false, rescode.Failed(err)
	}
	return res != "", nil
}

func (r *VerifyRedisRepo) Find(ctx context.Context, trc trace.Tracer, token string, deviceId string) (*auth.Verify, error) {
	ctx, span := trc.Start(ctx, "VerifyRedisRepo.Find")
	defer span.End()
	res, err := r.db.Get(ctx, r.calcKey(deviceId, token)).Result()
	if err != nil {
		return nil, rescode.Failed(err)
	}
	var e auth.Verify
	if err = json.Unmarshal([]byte(res), &e); err != nil {
		return nil, rescode.Failed(err)
	}
	return &e, nil
}

func (r *VerifyRedisRepo) Delete(ctx context.Context, trc trace.Tracer, token string, deviceId string) error {
	ctx, span := trc.Start(ctx, "VerifyRedisRepo.Delete")
	defer span.End()
	if err := r.db.Del(ctx, r.calcKey(deviceId, token)).Err(); err != nil {
		return rescode.Failed(err)
	}
	return nil
}

func (r *VerifyRedisRepo) calcKey(deviceId string, token string) string {
	return "verify" + "__" + token + "__" + deviceId
}
