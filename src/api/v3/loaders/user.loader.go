package loaders

import (
	"context"
	"time"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/v3/loaders"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/helpers"
	"github.com/SevenTV/GQL/src/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func userByID(gCtx global.Context) *loaders.UserLoader {
	return loaders.NewUserLoader(loaders.UserLoaderConfig{
		Wait: time.Millisecond * 5,
		Fetch: func(keys []primitive.ObjectID) ([]*model.User, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			// Fetch user data from the database
			models := make([]*model.User, len(keys))
			errs := make([]error, len(keys))

			// Initially fill the response with "deleted user" models in case some cannot be found
			deletedModel := helpers.UserStructureToModel(gCtx, structures.DeletedUser)
			for i := 0; i < len(models); i++ {
				models[i] = deletedModel
			}

			users, err := gCtx.Inst().Query.Users(ctx, bson.M{
				"_id": bson.M{
					"$in": keys,
				},
			}).Items()

			if err == nil {
				m := make(map[primitive.ObjectID]structures.User)
				for _, u := range users {
					m[u.ID] = u
				}

				for i, v := range keys {
					if x, ok := m[v]; ok {
						models[i] = helpers.UserStructureToModel(gCtx, x)
					}
				}
			}

			return models, errs
		},
	})
}
