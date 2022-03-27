package query

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/structures/v3/query"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/auth"
	"github.com/SevenTV/GQL/src/api/v3/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const INBOX_QUERY_LIMIT_MOST = 1000

func (r *Resolver) Inbox(ctx context.Context, userID primitive.ObjectID, afterIDArg *primitive.ObjectID, limitArg *int) ([]*model.InboxMessage, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, errors.ErrUnauthorized()
	}

	// Pagination
	afterID := primitive.NilObjectID
	if afterIDArg != nil {
		afterID = *afterIDArg
	}
	limit := 100
	if limitArg != nil {
		limit = *limitArg
		if limit > INBOX_QUERY_LIMIT_MOST {
			limit = INBOX_QUERY_LIMIT_MOST
		}
	}

	// Fetch target user
	user := r.Ctx.Inst().Query.Users(ctx, bson.M{"_id": userID}).First()
	if user == nil {
		return nil, errors.ErrUnknownUser()
	}

	messages := r.Ctx.Inst().Query.InboxMessages(ctx, query.InboxMessagesQueryOptions{
		Actor:               actor,
		User:                user,
		Limit:               limit,
		AfterID:             afterID,
		SkipPermissionCheck: false,
	})
	if messages.Error() != nil {
		return nil, messages.Error()
	}

	result := make([]*model.InboxMessage, len(messages.Items()))
	for i, msg := range messages.Items() {
		result[i] = helpers.MessageStructureToInboxModel(r.Ctx, msg)
	}

	return result, nil
}
