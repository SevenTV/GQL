package query

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/helpers"
	"github.com/SevenTV/GQL/src/api/v3/loaders"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const EMOTES_QUERY_LIMIT = 300

func (r *Resolver) Emote(ctx context.Context, id primitive.ObjectID) (*model.Emote, error) {
	emote, err := loaders.For(ctx).EmoteByID.Load(id)
	if emote == nil || emote.ID == structures.DeletedEmote.ID {
		return nil, errors.ErrUnknownEmote()
	}
	return emote, err
}

func (r *Resolver) Emotes(ctx context.Context, query string, pageArg *int, limitArg *int, filterArg *model.EmoteSearchFilter, sortArg *model.Sort) (*model.EmoteSearchResult, error) {
	// Define limit (how many emotes can be returned in a single query)
	limit := 20
	if limitArg != nil {
		limit = *limitArg
	}
	if limit > EMOTES_QUERY_LIMIT {
		limit = EMOTES_QUERY_LIMIT
	}

	// Define default filter
	filter := filterArg
	if filter == nil {
		filter = &model.EmoteSearchFilter{
			CaseSensitive: utils.BoolPointer(false),
			ExactMatch:    utils.BoolPointer(false),
		}
	} else {
		filter = filterArg
	}

	// Define the query string
	query = strings.Trim(query, " ")

	// Retrieve pagination values
	page := 1
	if pageArg != nil {
		page = *pageArg
	}
	if page < 1 {
		page = 1
	}

	// Retrieve sorting options
	sortopt := &model.Sort{
		Value: "popularity",
		Order: model.SortOrderAscending,
	}
	if sortArg != nil {
		sortopt = sortArg
	}

	// Set up db query
	match := bson.D{{Key: "versions.0.state.lifecycle", Value: structures.EmoteLifecycleLive}}

	// Define the pipeline
	pipeline := mongo.Pipeline{}

	// Define sorting
	// (will be ignored in the case of exact search)
	order, validOrder := sortOrderMap[string(sortopt.Order)]
	field, validField := sortFieldMap[sortopt.Value]

	// Apply name/tag query
	h := sha256.New()
	h.Write(utils.S2B(query))
	queryKey := r.Ctx.Inst().Redis.ComposeKey("gql-v3", fmt.Sprintf("emote-search:%s", hex.EncodeToString((h.Sum(nil)))))
	cpargs := bson.A{}

	// Handle exact match
	if filter.ExactMatch != nil && *filter.ExactMatch {
		// For an exact mathc we will use the $text operator
		// rather than $indexOfCP because name/tags are indexed fields
		match = append(match, bson.E{Key: "$text", Value: bson.M{
			"$search":        query,
			"$caseSensitive": filter.CaseSensitive != nil && *filter.CaseSensitive,
		}})
		pipeline = append(pipeline, []bson.D{
			{{Key: "$match", Value: match}},
			{{Key: "$sort", Value: bson.M{"score": bson.M{"$meta": "textScore"}}}},
		}...)
	} else {
		or := bson.A{}
		if filter.CaseSensitive != nil && *filter.CaseSensitive {
			cpargs = append(cpargs, "$name", query)
		} else {
			cpargs = append(cpargs, bson.M{"$toLower": "$name"}, strings.ToLower(query))
		}

		or = append(or, bson.M{
			"$expr": bson.M{
				"$gt": bson.A{bson.M{"$indexOfCP": cpargs}, -1},
			},
		})

		// Add tag search
		if filter.IgnoreTags == nil || !*filter.IgnoreTags {
			or = append(or, bson.M{
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
			})
		}

		match = append(match, bson.E{Key: "$or", Value: or})
		if validOrder && validField {
			pipeline = append(pipeline, bson.D{
				{Key: "$sort", Value: bson.M{field: order}},
			})
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// Complete the pipeline
	totalCount, countErr := r.Ctx.Inst().Redis.RawClient().Get(ctx, string(queryKey)).Int()
	wg := sync.WaitGroup{}
	wg.Add(1)
	if countErr == redis.Nil {
		go func() { // Run a separate pipeline to return the total count that could be paginated
			defer wg.Done()
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
			dur := utils.Ternary(query == "", time.Minute*10, time.Hour*1).(time.Duration)
			if err = r.Ctx.Inst().Redis.SetEX(ctx, queryKey, totalCount, dur); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"key":   queryKey,
					"count": totalCount,
				}).Error("redis, failed to save total list count of emotes() gql query")
			}
		}()
	} else {
		wg.Done()
	}

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
		// Sort by version timestamp
		sort.Slice(e.Versions, func(i, j int) bool {
			a := e.Versions[i]
			b := e.Versions[j]

			return b.Timestamp.Before(a.Timestamp)
		})

		// Bring forward the latest version
		if len(e.Versions) > 0 {
			e.ID = e.Versions[0].ID
		}
		models[i] = helpers.EmoteStructureToModel(r.Ctx, e)
	}

	return &model.EmoteSearchResult{
		Count: totalCount,
		Items: models,
	}, nil
}

var sortFieldMap = map[string]string{
	"age":        "_id",
	"popularity": "versions.state.channel_count",
}
