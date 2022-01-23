package cache

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/SevenTV/GQL/src/global"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type redisCache struct {
	gCtx   global.Context
	prefix string
	ttl    time.Duration
}

func NewRedisCache(ctx global.Context, prefix string, ttl time.Duration) graphql.Cache {
	return &redisCache{
		gCtx:   ctx,
		prefix: prefix,
		ttl:    ttl,
	}
}

func (c *redisCache) Get(ctx context.Context, key string) (value interface{}, ok bool) {
	v, err := c.gCtx.Inst().Redis.Get(ctx, c.prefix+key)
	if err == nil {
		return v, true
	}

	if err != redis.Nil {
		logrus.WithError(err).Error("failed to query redis")
	}

	return nil, false
}

func (c *redisCache) Add(ctx context.Context, key string, value interface{}) {
	err := c.gCtx.Inst().Redis.SetEX(ctx, c.prefix+key, value, c.ttl)
	if err != nil {
		logrus.WithError(err).Error("failed to query redis")
	}
}
