package query

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/utils"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const EMOTES_QUERY_LIMIT int32 = 150

var searchRegex = regexp.MustCompile(`[.*+?^${}()|[\\]\\\\]`)

func (r *Resolver) Emotes(ctx context.Context, args struct {
	Query    string
	Limit    *int32
	AfterID  *string
	BeforeID *string
	Sort     *Sort
}) ([]*EmoteResolver, error) {
	// Define limit (how many emotes can be returned in a single query)
	limit := int32(20)
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit > EMOTES_QUERY_LIMIT {
		limit = EMOTES_QUERY_LIMIT
	}

	// Define the query
	query := strings.Trim(args.Query, " ")
	lQuery := fmt.Sprintf("(?i)%s", strings.ToLower(searchRegex.ReplaceAllString(query, "\\\\$0")))

	//
	match := bson.M{
		"status": structures.EmoteStatusLive,
	}

	// Retrieve pagination values
	// AfterID is the ID of the emote to paginate from
	pagination := bson.M{}
	if args.AfterID != nil && *args.AfterID != "" {
		if afterID, err := primitive.ObjectIDFromHex(*args.AfterID); err != nil {
			return nil, err
		} else {
			pagination["$gt"] = afterID
		}
	}
	// BeforeID is the ID of the emote to paginate until
	if args.BeforeID != nil && *args.BeforeID != "" {
		if beforeID, err := primitive.ObjectIDFromHex(*args.BeforeID); err != nil {
			return nil, err
		} else {
			pagination["$lt"] = beforeID
		}
	}
	// Apply pagination if after or before was specified
	if len(pagination) > 0 {
		match["_id"] = pagination
	}

	// Apply name/tag query
	if len(query) > 0 {
		match["$or"] = bson.A{
			bson.M{
				"name": bson.M{
					"$regex": lQuery,
				},
			},
			bson.M{
				"tags": bson.M{
					"$regex": lQuery,
				},
			},
		}
	}

	// Define the pipeline
	pipeline := mongo.Pipeline{
		// Step 1: match the query
		bson.D{bson.E{Key: "$match", Value: match}},
	}

	// Handle sorting
	if args.Sort != nil {
		sort := args.Sort
		order, validOrder := sortOrderMap[string(sort.Order)]
		field, validField := sortFieldMap[sort.Value]

		if validOrder && validField {
			pipeline = append(pipeline, bson.D{bson.E{
				Key:   "$sort",
				Value: bson.D{bson.E{Key: field, Value: order}},
			}})
		}
	}

	// Complete the pipeline
	pipeline = append(pipeline, []bson.D{
		// Step 2: a faceted call, which simultaneously gets emotes and returns total query-scoped collection count
		{bson.E{
			Key: "$facet",
			Value: bson.M{
				"_count": []bson.M{{"$count": "value"}},
				"emotes": []bson.M{{"$limit": limit}},
			},
		}},
		// Remove the _count array value, replacing it by "count" as int
		{bson.E{Key: "$addFields", Value: bson.M{"count": bson.M{"$first": "$_count.value"}}}},
		{bson.E{Key: "$unset", Value: "_count"}},
	}...)

	// Begin the pipeline, fetching the emotes
	cur, err := mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, pipeline)
	if err != nil && err != mongo.ErrNoDocuments {
		log.WithError(err).Error("mongo")
		return nil, err
	}

	// Parse emotes to structs
	var result emotesPipelineResult
	cur.Next(ctx)
	if err := cur.Decode(&result); err != nil {
		log.WithError(err).Error("mongo")
	}
	cur.Close(ctx)

	// Create resolvers for the returned emotes
	emotes := result.Emotes
	resolvers := make([]*EmoteResolver, len(emotes))
	fields := GenerateSelectedFieldMap(ctx)
	for i, emote := range emotes {
		r, err := CreateEmoteResolver(ctx, emote, &emote.ID, fields.Children)
		if err != nil {
			return nil, err
		}

		resolvers[i] = r
	}

	// Send extra meta to be returned with the query
	// This contains the total amount of emotes seen in the query
	req := ctx.Value(utils.ReqKey).(*fiber.Ctx)
	req.Locals("meta", map[string]interface{}{
		"emotes": map[string]interface{}{
			"total": result.Count,
		},
	})

	return resolvers, nil
}

var sortFieldMap = map[string]string{
	"age":        "_id",
	"popularity": "channel_count",
}

type emotesPipelineResult struct {
	Count  int32               `bson:"count"`
	Emotes []*structures.Emote `bson:"emotes"`
}
