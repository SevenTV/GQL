package emote

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/helpers"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// Fetch emote sets that have this emote
	setIDs := []primitive.ObjectID{}

	// Ping redis for a cached value
	rKey := r.Ctx.Inst().Redis.ComposeKey("gql-v3", fmt.Sprintf("emote:%s:active_sets", obj.ID.Hex()))
	v, err := r.Ctx.Inst().Redis.Get(ctx, rKey)
	if err == nil && v != "" {
		if err = json.Unmarshal(utils.S2B(v), &setIDs); err != nil {
			logrus.WithError(err).Error("couldn't decode emote's active set ids")
		}
	} else {
		cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmoteSets).Find(ctx, bson.M{"emotes.id": obj.ID}, options.Find().SetProjection(bson.M{"owner_id": 1}))
		if err != nil {
			return nil, err
		}
		for i := 0; cur.Next(ctx); i++ {
			v := &structures.EmoteSet{}
			if err = cur.Decode(v); err != nil {
				logrus.WithError(err).Error("mongo, couldn't decode into EmoteSet")
			}
			setIDs = append(setIDs, v.ID)
		}

		// Set in redis
		b, err := json.Marshal(setIDs)
		if err = multierror.Append(err, r.Ctx.Inst().Redis.SetEX(ctx, rKey, utils.B2S(b), time.Hour*6)).ErrorOrNil(); err != nil {
			logrus.WithError(err).Error("failed to cache set ids in redis")
		}
	}

	// Fetch users with this set active
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, aggregations.Combine(
		mongo.Pipeline{
			{{
				Key:   "$sort",
				Value: bson.M{"metadata.role_position": -1},
			}},
			{{
				Key: "$match",
				Value: bson.M{
					"connections.emote_set_id": bson.M{
						"$in": setIDs,
					},
				},
			}},
			{{Key: "$skip", Value: (page - 1) * limit}},
			{{
				Key:   "$limit",
				Value: limit,
			}},
		},
		aggregations.UserRelationRoles,
	))
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
