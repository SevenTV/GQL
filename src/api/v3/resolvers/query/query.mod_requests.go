package query

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/query"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/auth"
	"github.com/SevenTV/GQL/src/api/v3/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModRequests implements generated.QueryResolver
func (r *Resolver) ModRequests(ctx context.Context, afterIDArg *primitive.ObjectID) ([]*model.ModRequestMessage, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, errors.ErrUnauthorized()
	}
	afterID := primitive.NilObjectID
	if afterIDArg != nil {
		afterID = *afterIDArg
	}

	match := bson.M{
		"kind": structures.MessageKindModRequest,
		"read": false,
	}
	if !afterID.IsZero() {
		match["_id"] = bson.M{"$gt": afterID}
	}
	messages, err := r.Ctx.Inst().Query.ModRequestMessages(ctx, query.ModRequestMessagesQueryOptions{
		Actor: actor,
		Targets: map[structures.ObjectKind]bool{
			structures.ObjectKindEmote: true,
		},
	})
	if err != nil {
		graphql.AddError(ctx, err)
		return []*model.ModRequestMessage{}, nil
	}

	result := make([]*model.ModRequestMessage, len(messages))
	for i, msg := range messages {
		result[i] = helpers.MessageStructureToModRequestModel(r.Ctx, msg)
	}
	return result, nil
}
