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

func emoteSetLoader(gCtx global.Context) *loaders.EmoteSetLoader {
	return loaders.NewEmoteSetLoader(loaders.EmoteSetLoaderConfig{
		Wait: time.Millisecond * 25,
		Fetch: func(keys []primitive.ObjectID) ([]*model.EmoteSet, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			// Fetch emote set data from the database
			models := make([]*model.EmoteSet, len(keys))
			errs := make([]error, len(keys))
			cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameEmoteSets).Aggregate(ctx, aggregations.Combine(
				mongo.Pipeline{
					{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": keys}}}},
					{{
						Key: "$lookup",
						Value: mongo.Lookup{
							From:         mongo.CollectionNameEmotes,
							LocalField:   "emotes.id",
							ForeignField: "_id",
							As:           "_emotes",
						},
					}},
					{{
						Key: "$set",
						Value: bson.M{
							"emotes": bson.M{"$map": bson.M{
								"input": "$emotes",
								"in": bson.M{"$mergeObjects": bson.A{
									"$$this",
									bson.M{"emote": bson.M{
										"$arrayElemAt": bson.A{"$_emotes", bson.M{"$indexOfArray": bson.A{"$_emotes._id", "$$this.id"}}},
									}},
								}},
							}},
						},
					}},
				},
				aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{Roles: true}),
			))
			if err != nil {
				logrus.WithError(err).Error("mongo, failed to spawn aggregation")
			}

			// Iterate over cursor
			// Transform emote set structures into models
			m := make(map[primitive.ObjectID]*structures.EmoteSet)
			for i := 0; cur.Next(ctx); i++ {
				v := &structures.EmoteSet{}
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
					models[i] = helpers.EmoteSetStructureToModel(gCtx, x)
				}
			}

			return models, errs
		},
	})
}
