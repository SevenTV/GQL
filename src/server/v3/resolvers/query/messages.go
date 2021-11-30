package query

import (
	"context"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) Inbox(ctx context.Context, args struct {
	AfterID *string
}) ([]*MessageResolver, error) {
	// Get the actor
	actor, ok := ctx.Value(helpers.UserKey).(*structures.User)
	if !ok {
		return nil, helpers.ErrAccessDenied
	}

	// Query
	match := bson.M{}
	pagination := bson.M{}
	if args.AfterID != nil && primitive.IsValidObjectID(*args.AfterID) {
		pagination["$gt"], _ = primitive.ObjectIDFromHex(*args.AfterID)
	}
	if len(pagination) > 0 {
		match["_id"] = pagination
	}

	// Create pipeline
	fields := GenerateSelectedFieldMap(ctx).Children
	pipeline := mongo.Pipeline{
		// Stage 1: find only readstates with the actor as recipient
		{{Key: "$match", Value: bson.M{"recipient_id": actor.ID}}},
		// Stage 2: Lookup the messages collection for message data
		{{
			Key: "$lookup",
			Value: mongo.LookupWithPipeline{
				From: mongo.CollectionNameMessages,
				Let:  bson.M{"msg_id": "$message_id", "read": "$read"},
				Pipeline: &mongo.Pipeline{
					// Stage 2.1: Filter to the message in the inner pipeline
					{{
						Key: "$match",
						Value: bson.M{
							"$expr": bson.M{
								"$eq": bson.A{"$$msg_id", "$_id"},
							},
						},
					}},
					// Stage 2.2: Add read boolean to the message's data
					{{Key: "$set", Value: bson.M{"read": "$$read"}}},
				},
				As: "_msg",
			},
		}},
		// Stage 3: swap in the readstate for a full message
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": bson.M{"$first": "$_msg"}}}},
		// Stage 4: filter to the user's query
		{{Key: "$match", Value: match}},
	}
	// Handle author relation
	if _, ok := fields["author"]; ok {
		pipeline = append(pipeline, []bson.D{
			{{
				Key: "$lookup",
				Value: mongo.LookupWithPipeline{
					From: mongo.CollectionNameUsers,
					Let:  bson.M{"author_id": "$author_id"},
					Pipeline: func() *mongo.Pipeline {
						p := mongo.Pipeline{
							{{
								Key: "$match",
								Value: bson.M{
									"$expr": bson.M{"$eq": bson.A{"$$author_id", "$_id"}},
								},
							}},
						}
						p = append(p, aggregations.UserRelationRoles...)
						return &p
					}(),
					As: "_author",
				},
			}},
			{{Key: "$set", Value: bson.M{"author": bson.M{"$first": "$_author"}}}}, {{Key: "$unset", Value: bson.A{"_author"}}},
		}...)
	}

	// Run aggregation
	messages := []*structures.Message{}
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameMessagesRead).Aggregate(ctx, pipeline)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	if err = cur.All(ctx, &messages); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Generate resolvers
	resolvers := make([]*MessageResolver, len(messages))
	for i, msg := range messages {
		resolver, err := CreateMessageResolver(r.Ctx, ctx, msg, &msg.ID, fields)
		if err != nil {
			logrus.WithError(err).Error("CreateMessageResolver")
			return nil, err
		}

		resolvers[i] = resolver
	}

	return resolvers, nil
}
