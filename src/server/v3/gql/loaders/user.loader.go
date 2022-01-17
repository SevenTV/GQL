package loaders

import (
	"time"

	"github.com/SevenTV/GQL/graph/loaders"
	"github.com/SevenTV/GQL/graph/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userLoader = loaders.NewUserLoader(loaders.UserLoaderConfig{
	Wait: time.Millisecond * 50,
	Fetch: func(keys []primitive.ObjectID) ([]*model.User, []error) {
		return nil, nil
	},
})
