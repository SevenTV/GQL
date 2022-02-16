package global

import (
	"github.com/SevenTV/Common/redis"
	"github.com/SevenTV/GQL/src/instance"
)

type Instances struct {
	Mongo instance.Mongo
	Redis redis.Instance
}
