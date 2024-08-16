package repository

import (
	"context"
	"encoding/json"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

type SessionRedisRepo struct {
	syncRepo
	db *redis.Client
}

func NewSessionRedisRepo(db *redis.Client) *SessionRedisRepo {
	return &SessionRedisRepo{
		db: db,
	}
}

func (s *SessionRedisRepo) Save(ctx context.Context, t trace.Tracer, opts auth.SessionSaveOpts) error {
	ctx, span := t.Start(ctx, "SessionRedisRepo.Save")
	defer span.End()
	s.syncRepo.Lock()
	defer s.syncRepo.Unlock()
	key := s.calcKey(opts.UserId, opts.Session.DeviceId)
	bytes, err := json.Marshal(opts.Session)
	if err != nil {
		return rescode.Failed(err)
	}
	if err := s.checkExistAndDel(ctx, t, key); err != nil {
		return rescode.Failed(err)
	}
	if err := s.db.Set(ctx, key, bytes, 0).Err(); err != nil {
		return rescode.Failed(err)
	}
	return nil
}

func (s *SessionRedisRepo) Find(ctx context.Context, t trace.Tracer, opts auth.SessionFindOpts) (*auth.Session, bool, error) {
	ctx, span := t.Start(ctx, "SessionRedisRepo.FindByIds")
	defer span.End()
	key := s.calcKey(opts.UserId, opts.DeviceId)
	e, notExists, err := s.getByKey(ctx, t, key)
	if err != nil {
		return nil, false, rescode.Failed(err)
	}
	if notExists {
		return nil, true, nil
	}
	return e, false, nil
}

func (s *SessionRedisRepo) FindAllByUser(ctx context.Context, t trace.Tracer, opts auth.FindAllByUserOpts) ([]*auth.Session, error) {
	ctx, span := t.Start(ctx, "SessionRedisRepo.FindAllByUserId")
	defer span.End()
	keys, err := s.db.Keys(ctx, s.calcKey(opts.UserId, "*")).Result()
	if err != nil {
		return nil, rescode.Failed(err)
	}
	entities := make([]*auth.Session, len(keys))
	for i, k := range keys {
		e, _, err := s.getByKey(ctx, t, k)
		if err != nil {
			return nil, rescode.Failed(err)
		}
		entities[i] = e
	}
	return entities, nil
}

func (s *SessionRedisRepo) checkExistAndDel(ctx context.Context, t trace.Tracer, key string) error {
	ctx, span := t.Start(ctx, "SessionRedisRepo.checkExistAndDel")
	defer span.End()
	exist, err := s.db.Exists(ctx, key).Result()
	if err != nil {
		return rescode.Failed(err)
	}
	if exist == 1 {
		return s.db.Del(ctx, key).Err()
	}
	return nil
}

func (s *SessionRedisRepo) calcKey(userId uuid.UUID, deviceId string) string {
	return deviceId + "__" + userId.String()
}

func (s *SessionRedisRepo) getByKey(ctx context.Context, t trace.Tracer, key string) (*auth.Session, bool, error) {
	ctx, span := t.Start(ctx, "SessionRedisRepo.getByKey")
	defer span.End()
	res, err := s.db.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, true, nil
		}
		return nil, true, rescode.Failed(err)
	}
	var e auth.Session
	if err := json.Unmarshal([]byte(res), &e); err != nil {
		return nil, false, rescode.Failed(err)
	}
	return &e, false, nil
}

func (s *SessionRedisRepo) Destroy(ctx context.Context, t trace.Tracer, opts auth.SessionDestroyOpts) error {
	ctx, span := t.Start(ctx, "SessionRedisRepo.Destroy")
	defer span.End()
	key := s.calcKey(opts.UserId, opts.DeviceId)
	if err := s.db.Del(ctx, key).Err(); err != nil {
		return rescode.Failed(err)
	}
	return nil
}
