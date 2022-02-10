package emote

import (
	"context"
	"strconv"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/helpers"
	"go.mongodb.org/mongo-driver/bson"
)

const EMOTE_CHANNEL_QUERY_SIZE_MOST = 25
const EMOTE_CHANNEL_QUERY_PAGE_CAP = 50

func (r *Resolver) Channels(ctx context.Context, obj *model.Emote, pageArg *int, limitArg *int) (*model.UserSearchResult, error) {
	limit := EMOTE_CHANNEL_QUERY_SIZE_MOST
	if limitArg != nil {
		limit = *limitArg
	}
	if limit > EMOTE_CHANNEL_QUERY_SIZE_MOST {
		limit = EMOTE_CHANNEL_QUERY_SIZE_MOST
	}
	page := 1
	if pageArg != nil {
		page = *pageArg
	}
	if page < 1 {
		page = 1
	}
	if page > EMOTE_CHANNEL_QUERY_PAGE_CAP {
		return nil, errors.ErrInvalidRequest().SetFields(errors.Fields{
			"PAGE":  strconv.Itoa(page),
			"LIMIT": strconv.Itoa(EMOTE_CHANNEL_QUERY_PAGE_CAP),
		}).SetDetail("No further pagination is allowed")
	}

	pipeline := mongo.Pipeline{
		{{
			Key:   "$sort",
			Value: bson.M{"metadata.role_position": -1},
		}},
		{{Key: "$skip", Value: (page - 1) * limit}},
		{{
			Key:   "$unwind",
			Value: bson.M{"path": "$connections"},
		}},
		{{
			Key: "$lookup",
			Value: mongo.LookupWithPipeline{
				From: mongo.CollectionNameEmoteSets,
				Let:  bson.M{"set_id": "$connections.emote_set_id"},
				Pipeline: &mongo.Pipeline{
					{{
						Key: "$match",
						Value: bson.M{
							"emotes.id": obj.ID,
							"$expr": bson.M{
								"$eq": bson.A{"$_id", "$$set_id"},
							},
						},
					}},
					{{Key: "$project", Value: bson.M{"_id": 1}}},
				},
				As: "sets",
			},
		}},
		{{Key: "$unset", Value: "connections"}},
		{{
			Key: "$set",
			Value: bson.M{"ok": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$gt": bson.A{bson.M{"$size": "$sets"}, 0}},
					"then": true,
					"else": false,
				},
			}},
		}},
		{{Key: "$match", Value: bson.M{"ok": true}}},
		{{Key: "$limit", Value: limit}},
	}

	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, aggregations.Combine(pipeline, aggregations.UserRelationRoles))
	if err != nil {
		return nil, err
	}

	users := []*structures.User{}
	if err = cur.All(ctx, &users); err != nil {
		return nil, err
	}

	models := make([]*model.User, len(users))
	for i, u := range users {
		if u.ID.IsZero() {
			u = structures.DeletedUser
		}
		models[i] = helpers.UserStructureToModel(r.Ctx, u)
	}
	results := model.UserSearchResult{
		Count: 0,
		Items: models,
	}
	return &results, nil
}
