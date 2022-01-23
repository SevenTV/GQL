package instance

import (
	"context"
	"time"

	"github.com/SevenTV/Common/redis"
	rawRedis "github.com/go-redis/redis/v8"
)

type Redis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
	SetEX(ctx context.Context, key string, value interface{}, expiry time.Duration) error
	Exists(ctx context.Context, keys ...string) (int, error)
	IncrBy(ctx context.Context, key string, amount int) (int, error)
	DecrBy(ctx context.Context, key string, amount int) (int, error)
	Expire(ctx context.Context, key string, expiry time.Duration) error
	Del(ctx context.Context, keys ...string) (int, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Pipeline(ctx context.Context) rawRedis.Pipeliner

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

func (r *redisInst) Exists(ctx context.Context, keys ...string) (int, error) {
	i, err := r.Instance.RawClient().Exists(ctx, keys...).Result()
	return int(i), err
}

func (r *redisInst) IncrBy(ctx context.Context, key string, amount int) (int, error) {
	i, err := r.Instance.RawClient().IncrBy(ctx, key, int64(amount)).Result()
	return int(i), err
}

func (r *redisInst) DecrBy(ctx context.Context, key string, amount int) (int, error) {
	i, err := r.Instance.RawClient().DecrBy(ctx, key, int64(amount)).Result()
	return int(i), err
}

func (r *redisInst) Expire(ctx context.Context, key string, expiry time.Duration) error {
	return r.Instance.RawClient().Expire(ctx, key, expiry).Err()
}

func (r *redisInst) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Instance.RawClient().TTL(ctx, key).Result()
}

func (r *redisInst) Del(ctx context.Context, keys ...string) (int, error) {
	i, err := r.Instance.RawClient().Del(ctx, keys...).Result()
	return int(i), err
}

func (r *redisInst) Pipeline(ctx context.Context) rawRedis.Pipeliner {
	return r.Instance.RawClient().Pipeline()
}

const RedisPrefix = "api-v3:gql"

func WrapRedis(redis redis.Instance) Redis {
	return &redisInst{Instance: redis}
}
