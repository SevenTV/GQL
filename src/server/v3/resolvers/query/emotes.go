package query

import (
	"context"
	"strings"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const EMOTES_QUERY_LIMIT int32 = 300

func (r *Resolver) Emotes(ctx context.Context, args struct {
	Query string
	Page  *int32
	Limit *int32
	Sort  *Sort
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

	//
	match := bson.M{
		"status": structures.EmoteStatusLive,
	}

	// Retrieve pagination values
	page := int32(1)
	if args.Page != nil {
		page = *args.Page
		if page < 1 {
			page = 1
		}
	}

	// Apply name/tag query
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
	pipeline := mongo.Pipeline{}

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

	// Create a sub-pipeline for emotes
	// This is to fetch relational data in advance
	emoteSubPipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
	}
	// Add owner data?

	{
		fields := GenerateSelectedFieldMap(ctx).Children
		if _, ok := fields["owner"]; ok {
			o := fields["owner"]

			_, qEditors := o.Children["editors"]
			_, qRoles := o.Children["roles"]
			if !qRoles {
				_, qRoles = o.Children["tag_color"]
			}
			_, qChannelEmotes := o.Children["channel_emotes"]
			emoteSubPipeline = append(emoteSubPipeline, aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{
				Editors:       qEditors,
				Roles:         qRoles,
				ChannelEmotes: qChannelEmotes,
			})...)
		}
	}

	// Complete the pipeline
	pipeline = append(pipeline, []bson.D{
		{{
			Key: "$facet",
			Value: bson.M{
				"emotes": append(emoteSubPipeline, []bson.D{
					{{
						Key:   "$skip",
						Value: (page - 1) * limit,
					}},
					{{
						Key:   "$limit",
						Value: limit,
					}},
				}...),
				"count": []bson.M{
					{"$match": match},
					{"$count": "value"},
				},
			},
		}},
		{{
			Key:   "$set",
			Value: bson.M{"count": bson.M{"$first": "$count.value"}},
		}},
	}...)

	// Begin the pipeline, fetching the emotes
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, pipeline)
	if err != nil && err != mongo.ErrNoDocuments {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Parse emotes to structs
	result := emotesPipelineResult{}
	cur.Next(ctx)
	if err := cur.Decode(&result); err != nil {
		logrus.WithError(err).Error("mongo")
	}
	cur.Close(ctx)

	// Create resolvers for the returned emotes
	emotes := result.Emotes
	resolvers := []*EmoteResolver{}
	fields := GenerateSelectedFieldMap(ctx)
	for _, emote := range emotes {
		if emote == nil {
			continue
		}
		if emote.Owner == nil {
			emote.OwnerID = structures.DeletedUser.ID
			emote.Owner = structures.DeletedUser
		}

		r, err := CreateEmoteResolver(r.Ctx, ctx, emote, &emote.ID, fields.Children)
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, r)
	}

	// Send extra meta to be returned with the query
	// This contains the total amount of emotes seen in the query
	req := ctx.Value(utils.Key("request")).(*fiber.Ctx)
	req.Locals("meta", map[string]interface{}{
		"emotes": map[string]interface{}{
			"count": result.Count,
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
