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

func userByID(gCtx global.Context) *loaders.UserLoader {
	return loaders.NewUserLoader(loaders.UserLoaderConfig{
		Wait: time.Millisecond * 25,
		Fetch: func(keys []string) ([]*model.User, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			models := make([]*model.User, len(keys))
			errs := make([]error, len(keys))

			// Parse IDs
			ids := make([]primitive.ObjectID, len(keys))
			for i, k := range keys {
				id, err := primitive.ObjectIDFromHex(k)
				if err != nil {
					errs[i] = err
					continue
				}
				ids[i] = id
			}

			// Fetch users
			users, err := gCtx.Inst().Query.Users(ctx, bson.M{
				"_id": bson.M{"$in": ids},
			})
			if err == nil {
				m := make(map[primitive.ObjectID]*structures.User)
				for _, u := range users {
					if u == nil {
						continue
					}
					m[u.ID] = u
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
