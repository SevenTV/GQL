package loaders

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/GQL/graph/loaders"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/global"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func emoteLoader(gCtx global.Context) *loaders.EmoteLoader {
	return loaders.NewEmoteLoader(loaders.EmoteLoaderConfig{
		Fetch: func(keys []primitive.ObjectID) ([]*model.Emote, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			// Fetch emote data from the database
			models := make([]*model.Emote, len(keys))
			errs := make([]error, len(keys))
			cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, aggregations.Combine(
				mongo.Pipeline{{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": keys}}}}},
				aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{Roles: true}),
			))
			if err != nil {
				logrus.New().WithError(err).Error("mongo, failed to spawn aggregation")
			}

			// Initially fill the response with unknown emotes in case some cannot be found
			unknownModel := helpers.EmoteStructureToModel(gCtx, structures.DeletedEmote)
			for i := 0; i < len(models); i++ {
				models[i] = unknownModel
			}
			// Iterate over cursor
			// Transform emote structures into models
			m := make(map[primitive.ObjectID]*structures.Emote)
			for i := 0; cur.Next(ctx); i++ {
				v := &structures.Emote{}
				if err = cur.Decode(v); err != nil {
					errs[i] = err
				}
				m[v.ID] = v
			}
			if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
				logrus.WithError(err).Error("mongo, failed to close the cursor")
			}

			for i, v := range keys {
				if x, ok := m[v]; ok {
					models[i] = helpers.EmoteStructureToModel(gCtx, x)
				}
			}

			return models, errs
		},
		Wait: time.Millisecond * 5,
	})
}
