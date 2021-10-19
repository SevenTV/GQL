package instance

import "github.com/SevenTV/Common/redis"

type Redis interface {
	redis.Instance
}

const RedisPrefix = "api-v3:gql"
