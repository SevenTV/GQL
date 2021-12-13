package instance

import (
	"context"
	"time"

	"github.com/SevenTV/Common/redis"
)

type Redis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
	SetEX(ctx context.Context, key string, value interface{}, expiry time.Duration) error

	redis.Instance
}

type redisInst struct {
	redis.Instance
}

func (r *redisInst) Get(ctx context.Context, key string) (string, error) {
	return r.Instance.RawClient().Get(ctx, key).Result()
}

func (r *redisInst) Set(ctx context.Context, key string, value interface{}) error {
	return r.Instance.RawClient().Set(ctx, key, value, 0).Err()
}

func (r *redisInst) SetEX(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return r.Instance.RawClient().SetEX(ctx, key, value, expiry).Err()
}

const RedisPrefix = "api-v3:gql"

func WrapRedis(redis redis.Instance) Redis {
	return &redisInst{Instance: redis}
}
