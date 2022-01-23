package loaders

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/GQL/graph/loaders"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/gql/helpers"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func userLoader(gCtx global.Context) *loaders.UserLoader {
	return loaders.NewUserLoader(loaders.UserLoaderConfig{
		Wait: time.Millisecond * 5,
		Fetch: func(keys []primitive.ObjectID) ([]*model.User, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			// Fetch user data from the database
			models := make([]*model.User, len(keys))
			errs := make([]error, len(keys))
			cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, aggregations.Combine(
				mongo.Pipeline{{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": keys}}}}},
				aggregations.UserRelationRoles,
				aggregations.UserRelationEditors,
				aggregations.UserRelationOwnedEmotes,
			))
			if err != nil {
				logrus.WithError(err).Error("mongo, failed to spawn aggregation")
			}

			// Initially fill the response with "deleted user" models in case some cannot be found
			deletedModel := helpers.UserStructureToModel(gCtx, structures.DeletedUser)
			for i := 0; i < len(models); i++ {
				models[i] = deletedModel
			}
			// Iterate over cursor
			// Transform user structures into models
			for i := 0; cur.TryNext(ctx); i++ {
				v := &structures.User{}
				if err = cur.Decode(v); err != nil {
					errs[i] = err
				}
				models[i] = helpers.UserStructureToModel(gCtx, v)
			}
			if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
				logrus.WithError(err).Error("mongo, failed to close the cursor")
			}

			return models, errs
		},
	})
}
