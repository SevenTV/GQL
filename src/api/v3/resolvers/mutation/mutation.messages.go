package mutation

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/mutations"
	"github.com/SevenTV/GQL/src/api/v3/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) ReadMessages(ctx context.Context, messageIds []primitive.ObjectID, read bool) (int, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return 0, errors.ErrUnauthorized()
	}

	// Fetch messages
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameMessages).Find(ctx, bson.M{
		"_id": bson.M{"$in": messageIds},
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, errors.ErrUnknownMessage().SetDetail("No messages found")
		}
		return 0, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Mutate messages
	messages := []*structures.Message{}
	if err := cur.All(ctx, &messages); err != nil {
		return 0, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	updated := 0
	for _, msg := range messages {
		result, err := r.Ctx.Inst().Mutate.SetMessageReadStates(ctx, structures.NewMessageBuilder(msg), read, mutations.MessageReadStateOptions{
			Actor:               actor,
			SkipPermissionCheck: false,
		})
		if result != nil {
			for _, er := range result.Errors {
				graphql.AddError(ctx, er)
			}
		}
		if err != nil {
			return 0, err
		}
		updated += int(result.Updated)
	}
	return updated, nil
}
