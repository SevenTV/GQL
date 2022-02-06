package query

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const EMOTES_QUERY_LIMIT = 300

func (r *Resolver) Emote(ctx context.Context, id primitive.ObjectID) (*model.Emote, error) {
	return loaders.For(ctx).EmoteByID.Load(id)
}

func (r *Resolver) Emotes(ctx context.Context, query string, pageArg *int, limitArg *int, filter *model.EmoteSearchFilter, sortArg *model.Sort) (*model.EmoteSearchResult, error) {
	// Define limit (how many emotes can be returned in a single query)
	limit := 20
	if limitArg != nil {
		limit = *limitArg
	}
	if limit > EMOTES_QUERY_LIMIT {
		limit = EMOTES_QUERY_LIMIT
	}

	// Define the query string
	query = strings.Trim(query, " ")

	// Set up db query
	match := bson.M{"state.lifecycle": structures.EmoteLifecycleLive}

	// Retrieve pagination values
	page := 1
	if pageArg != nil {
		page = *pageArg
	}
	if page < 1 {
		page = 1
	}

	// Apply name/tag query
	h := sha256.New()
	h.Write(utils.S2B(query))
	queryKey := r.Ctx.Inst().Redis.ComposeKey("gql-v3", hex.EncodeToString((h.Sum(nil))))
	if len(query) > 0 {
		match["$or"] = bson.A{
			bson.M{
				"$expr": bson.M{
					"$gt": bson.A{
						bson.M{"$indexOfCP": bson.A{bson.M{"$toLower": "$name"}, strings.ToLower(query)}},
						-1,
					},
				},
			},
			bson.M{
				"$expr": bson.M{
					"$gt": bson.A{
						bson.M{"$indexOfCP": bson.A{bson.M{"$reduce": bson.M{
							"input":        "$tags",
							"initialValue": " ",
							"in":           bson.M{"$concat": bson.A{"$$value", "$$this"}},
						}}, strings.ToLower(query)}},
						-1,
					},
				},
			},
		}
	}

	// Define the pipeline
	pipeline := mongo.Pipeline{{{Key: "$match", Value: match}}}

	// Handle sorting
	if sortArg != nil {
		sort := *sortArg
		order, validOrder := sortOrderMap[string(sort.Order)]
		field, validField := sortFieldMap[sort.Value]

		if validOrder && validField {
			pipeline = append(pipeline, bson.D{{
				Key:   "$sort",
				Value: bson.D{{Key: field, Value: order}},
			}})
		}
	}

	// Complete the pipeline
	wg := sync.WaitGroup{}
	wg.Add(1)
	totalCount := 0
	go func() { // Run a separate pipeline to return the total count that could be paginated
		defer wg.Done()

		val, _ := r.Ctx.Inst().Redis.Get(ctx, queryKey)
		if val != "" {
			totalCount, _ = strconv.Atoi(val)
			return
		}

		cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, aggregations.Combine(
			pipeline,
			mongo.Pipeline{
				{{Key: "$count", Value: "count"}},
			}),
		)
		result := make(map[string]int, 1)
		if err == nil {
			cur.Next(ctx)
			if err = multierror.Append(cur.Decode(&result), cur.Close(ctx)).ErrorOrNil(); err != nil {
				logrus.WithError(err).Error("mongo, couldn't count")
			}
		}

		// Return total count & cache
		totalCount = result["count"]
		if err = r.Ctx.Inst().Redis.SetEX(ctx, queryKey, totalCount, time.Minute*2); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"key":   queryKey,
				"count": totalCount,
			}).Error("redis, failed to save total list count of emotes() gql query")
		}
	}()

	// Paginate and fetch the relevant emotes
	result := []*structures.Emote{}
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, aggregations.Combine(
		pipeline,
		mongo.Pipeline{
			{{Key: "$skip", Value: (page - 1) * limit}},
			{{Key: "$limit", Value: limit}},
		},
		aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{Roles: true}),
	))
	if err == nil {
		if err = cur.All(ctx, &result); err != nil {
			logrus.WithError(err).Error("mongo, failed to fetch emotes")
		}
	}
	wg.Wait() // wait for total count to finish

	models := make([]*model.Emote, len(result))
	for i, e := range result {
		models[i] = helpers.EmoteStructureToModel(r.Ctx, e)
	}

	return &model.EmoteSearchResult{
		Count: totalCount,
		Items: models,
	}, nil
}

var sortFieldMap = map[string]string{
	"age":        "_id",
	"popularity": "channel_count",
}
