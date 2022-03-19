package loaders

import (
	"context"
	"time"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/v2/loaders"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/helpers"
	"github.com/SevenTV/GQL/src/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func userLoader(gCtx global.Context, keyName string) *loaders.UserLoader {
	return loaders.NewUserLoader(loaders.UserLoaderConfig{
		Wait: time.Millisecond * 25,
		Fetch: func(keys []string) ([]*model.User, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			models := make([]*model.User, len(keys))
			errs := make([]error, len(keys))

			// Parse IDs
			ids := make([]interface{}, len(keys))
			for i, k := range keys {
				id, err := primitive.ObjectIDFromHex(k)
				if err != nil {
					ids[i] = k
					continue
				}
				ids[i] = id
			}

			// Initially fill the response with "deleted user" models in case some cannot be found
			deletedModel := helpers.UserStructureToModel(gCtx, structures.DeletedUser)
			for i := 0; i < len(models); i++ {
				models[i] = deletedModel
			}

			// Fetch users
			users, _, err := gCtx.Inst().Query.SearchUsers(ctx, bson.M{
				keyName: bson.M{"$in": ids},
			})
			if err == nil {
				m := make(map[interface{}]*structures.User)
				for _, u := range users {
					if u == nil {
						continue
					}
					switch keyName {
					case "username":
						m[u.Username] = u
					default:
						m[u.ID] = u
					}
				}

				for i, v := range ids {
					if x, ok := m[v]; ok {
						models[i] = helpers.UserStructureToModel(gCtx, x)
					}
				}
			}

			return models, errs
		},
	})
}
