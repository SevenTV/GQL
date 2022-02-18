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

func emoteSetByID(gCtx global.Context) *loaders.EmoteSetLoader {
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
						Key: "$group",
						Value: bson.M{
							"_id":  nil,
							"sets": bson.M{"$push": "$$ROOT"},
						},
					}},
					{{
						Key: "$lookup",
						Value: mongo.Lookup{
							From:         mongo.CollectionNameEmotes,
							LocalField:   "sets.emotes.id",
							ForeignField: "_id",
							As:           "emotes",
						},
					}},
					{{
						Key: "$lookup",
						Value: mongo.Lookup{
							From:         mongo.CollectionNameUsers,
							LocalField:   "emotes.owner_id",
							ForeignField: "_id",
							As:           "emote_owners",
						},
					}},
				},
			))
			if err != nil {
				logrus.WithError(err).Error("mongo, failed to spawn aggregation")
			}

			// Get roles (to assign to emote owners)
			roles, _ := gCtx.Inst().Query.Roles(ctx, bson.M{})
			roleMap := make(map[primitive.ObjectID]*structures.Role)
			for _, role := range roles {
				roleMap[role.ID] = role
			}

			// Iterate over cursor
			// Transform emote set structures into models
			m := make(map[primitive.ObjectID]*structures.EmoteSet)
			for i := 0; cur.Next(ctx); i++ {
				v := &aggregatedEmoteSetByID{}
				if err = cur.Decode(v); err != nil {
					errs[i] = err
				}

				emoteMap := make(map[primitive.ObjectID]*structures.Emote)
				ownerMap := make(map[primitive.ObjectID]*structures.User)
				for _, emote := range v.Emotes {
					emoteMap[emote.ID] = emote
				}
				for _, user := range v.EmoteOwners {
					ownerMap[user.ID] = user
				}

				var ok bool
				for _, set := range v.Sets {
					for _, ae := range set.Emotes {
						if ae.Emote, ok = emoteMap[ae.ID]; ok {
							if ae.Emote.Owner, ok = ownerMap[ae.Emote.OwnerID]; ok {
								for _, roleID := range ae.Emote.Owner.RoleIDs {
									role, roleOK := roleMap[roleID]
									if !roleOK {
										continue
									}
									ae.Emote.Owner.Roles = append(ae.Emote.Owner.Roles, role)
								}
							}
						}
					}
					m[set.ID] = set
				}
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

type aggregatedEmoteSetByID struct {
	Sets        []*structures.EmoteSet `bson:"sets"`
	Emotes      []*structures.Emote    `bson:"emotes"`
	EmoteOwners []*structures.User     `bson:"emote_owners"`
}

func emoteSetByUserID(gCtx global.Context) *loaders.BatchEmoteSetLoader {
	return loaders.NewBatchEmoteSetLoader(loaders.BatchEmoteSetLoaderConfig{
		Wait: time.Millisecond * 25,
		Fetch: func(keys []primitive.ObjectID) ([][]*model.EmoteSet, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			// Fetch emote sets
			modelLists := make([][]*model.EmoteSet, len(keys))
			errs := make([]error, len(keys))
			cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameEmoteSets).Aggregate(ctx, aggregations.Combine(
				mongo.Pipeline{
					{{
						Key:   "$match",
						Value: bson.M{"owner_id": bson.M{"$in": keys}},
					}},
					{{
						Key: "$group",
						Value: bson.M{
							"_id": "$owner_id",
							"sets": bson.M{
								"$push": "$$ROOT",
							},
						},
					}},
					{{
						Key: "$lookup",
						Value: mongo.Lookup{
							From:         mongo.CollectionNameEmotes,
							LocalField:   "sets.emotes.id",
							ForeignField: "_id",
							As:           "emotes",
						},
					}},
					{{
						Key: "$lookup",
						Value: mongo.Lookup{
							From:         mongo.CollectionNameUsers,
							LocalField:   "emotes.owner_id",
							ForeignField: "_id",
							As:           "emote_owners",
						},
					}},
				},
			))
			if err != nil {
				logrus.WithError(err).Error("mongo, failed to spawn aggregation")
			}

			// Get roles (to assign to emote owners)
			roles, _ := gCtx.Inst().Query.Roles(ctx, bson.M{})
			roleMap := make(map[primitive.ObjectID]*structures.Role)
			for _, role := range roles {
				roleMap[role.ID] = role
			}

			// Iterate over cursor
			m := make(map[primitive.ObjectID][]*structures.EmoteSet)
			for i := 0; cur.Next(ctx); i++ {
				v := &aggregatedEmoteSetByUserID{}
				if err = cur.Decode(v); err != nil {
					errs[i] = err
				}

				// Map emotes bound to the set
				emoteMap := make(map[primitive.ObjectID]*structures.Emote)
				ownerMap := make(map[primitive.ObjectID]*structures.User)
				for _, emote := range v.Emotes {
					emoteMap[emote.ID] = emote
				}
				for _, user := range v.EmoteOwners {
					ownerMap[user.ID] = user
				}

				var ok bool
				for _, set := range v.Sets {
					for _, ae := range set.Emotes {
						if ae.Emote, ok = emoteMap[ae.ID]; ok {
							if ae.Emote.Owner, ok = ownerMap[ae.Emote.OwnerID]; ok {
								for _, roleID := range ae.Emote.Owner.RoleIDs {
									role, roleOK := roleMap[roleID]
									if !roleOK {
										continue
									}
									ae.Emote.Owner.Roles = append(ae.Emote.Owner.Roles, role)
								}
							}
						}
					}
				}
				m[v.UserID] = v.Sets
			}
			if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
				logrus.WithError(err).Error("mongo, failed to close the cursor")
			}

			for i, v := range keys {
				if x, ok := m[v]; ok {
					models := make([]*model.EmoteSet, len(x))
					for ii, set := range x {
						models[ii] = helpers.EmoteSetStructureToModel(gCtx, set)
					}
					modelLists[i] = models
				}
			}

			return modelLists, errs
		},
	})
}

type aggregatedEmoteSetByUserID struct {
	UserID      primitive.ObjectID     `bson:"_id"`
	Sets        []*structures.EmoteSet `bson:"sets"`
	Emotes      []*structures.Emote    `bson:"emotes"`
	EmoteOwners []*structures.User     `bson:"emote_owners"`
}
