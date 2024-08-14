package repository

import (
	"context"
	"encoding/json"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

func (s *SessionRedisRepo) Save(ctx context.Context, userId uuid.UUID, session *auth.Session) error {
	s.syncRepo.Lock()
	defer s.syncRepo.Unlock()
	key := s.calcKey(userId, session.DeviceId)
	bytes, err := json.Marshal(session)
	if err != nil {
		return rescode.Failed(err)
	}
	if err := s.checkExistAndDel(ctx, key); err != nil {
		return rescode.Failed(err)
	}
	if err := s.db.Set(ctx, key, bytes, 0).Err(); err != nil {
		return rescode.Failed(err)
	}
	return nil
}

func (s *SessionRedisRepo) FindByIds(ctx context.Context, userId uuid.UUID, deviceId string) (*auth.Session, bool, error) {
	key := s.calcKey(userId, deviceId)
	e, notExists, err := s.getByKey(ctx, key)
	if err != nil {
		return nil, false, rescode.Failed(err)
	}
	if notExists {
		return nil, true, nil
	}
	return e, false, nil
}

func (s *SessionRedisRepo) FindAllByUserId(ctx context.Context, userId uuid.UUID) ([]*auth.Session, error) {
	keys, err := s.db.Keys(ctx, s.calcKey(userId, "*")).Result()
	if err != nil {
		return nil, rescode.Failed(err)
	}
	entities := make([]*auth.Session, len(keys))
	for i, k := range keys {
		e, _, err := s.getByKey(ctx, k)
		if err != nil {
			return nil, rescode.Failed(err)
		}
		entities[i] = e
	}
	return entities, nil
}

func (s *SessionRedisRepo) checkExistAndDel(ctx context.Context, key string) error {
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

func (s *SessionRedisRepo) getByKey(ctx context.Context, key string) (*auth.Session, bool, error) {
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

func (s *SessionRedisRepo) Destroy(ctx context.Context, userId uuid.UUID, deviceId string) error {
	key := s.calcKey(userId, deviceId)
	if err := s.db.Del(ctx, key).Err(); err != nil {
		return rescode.Failed(err)
	}
	return nil
}
