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

var userLoader = func(gCtx global.Context) *loaders.UserLoader {
	return loaders.NewUserLoader(loaders.UserLoaderConfig{
		Wait: time.Millisecond * 50,
		Fetch: func(keys []primitive.ObjectID) ([]*model.User, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			// Fetch user data from the database
			models := make([]*model.User, len(keys))
			errs := make([]error, len(keys))
			cur, err := gCtx.Inst().Mongo.Collection(structures.CollectionNameUsers).Aggregate(ctx, append(
				mongo.Pipeline{{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": keys}}}}},
				aggregations.UserRelationRoles...,
			))
			if err != nil {
				logrus.WithError(err).Error("mongo, failed to spawn aggregation")
			}

			// Iterate over cursor
			// Transform user structures into models
			for i := 0; cur.TryNext(ctx); i++ {
				v := &structures.User{}
				if err = cur.Decode(v); err != nil {
					errs[i] = err
					continue
				}

				models[i] = helpers.UserStructureToModel(v)
			}
			if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
				logrus.WithError(err).Error("mongo, failed to close the cursor")
				return models, errs
			}

			return models, nil
		},
	})
}
