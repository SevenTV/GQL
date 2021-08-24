package query

import (
	"context"
	"regexp"

	"github.com/SevenTV/ThreeLetterAPI/src/mongo"
	"github.com/SevenTV/ThreeLetterAPI/src/structures"
	"github.com/SevenTV/ThreeLetterAPI/src/utils"
	"github.com/gofiber/fiber/v2"
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
	// query := strings.Trim(args.Query, " ")
	// lQuery := fmt.Sprintf("(?i)%s", strings.ToLower(searchRegex.ReplaceAllString(query, "\\\\$0")))

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

	// Define a MongoDB Pipeline
	pipeline := mongo.Pipeline{
		bson.D{bson.E{Key: "$match", Value: match}},
		bson.D{bson.E{Key: "$limit", Value: limit}},
	}

	// Begin the pipeline, fetching the emotes
	cur, err := mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, pipeline)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Parse emotes to structs
	var emotes []*structures.Emote
	if err := cur.All(ctx, &emotes); err != nil {
		return nil, err
	}

	// Create resolvers for the returned emotes
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
			"total": 123,
		},
	})

	return resolvers, nil
}
