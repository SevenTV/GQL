package mutation

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/structures/mutations"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) ReadMessages(ctx context.Context, args struct {
	MessageIDs []string
	Read       *bool
}) (*int32, error) {
	// Get the actor user
	actor, ok := ctx.Value(helpers.UserKey).(*structures.User)
	if !ok {
		return nil, helpers.ErrAccessDenied
	}

	// Parse message IDs
	messageIDs := []primitive.ObjectID{}
	for _, s := range args.MessageIDs {
		if !primitive.IsValidObjectID(s) {
			continue
		}
		id, _ := primitive.ObjectIDFromHex(s)
		messageIDs = append(messageIDs, id)
	}

	// Update the readstates
	value := true
	if args.Read != nil && !*args.Read {
		value = false
	}
	result, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameMessagesRead).UpdateMany(ctx, bson.M{
		"recipient_id": actor.ID,
		"message_id":   bson.M{"$in": messageIDs},
	}, bson.M{
		"$set": bson.M{"read": value},
	})
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	return utils.Int32Pointer(int32(result.ModifiedCount)), nil
}

func (r *Resolver) SendInboxMessage(ctx context.Context, args struct {
	Recipients []string
	Subject    string
	Content    string
	Important  *bool
	Anonymous  *bool
}) (*query.MessageResolver, error) {
	// Get the actor user
	actor, ok := ctx.Value(helpers.UserKey).(*structures.User)
	if !ok {
		return nil, helpers.ErrAccessDenied
	}
	// The actor must be allowed the "Send Messages" permission
	if !actor.HasPermission(structures.RolePermissionSendMessages) {
		return nil, helpers.ErrAccessDenied
	}

	// Parse recipient IDs
	recipientIDs := []primitive.ObjectID{}
	for _, s := range args.Recipients {
		if !primitive.IsValidObjectID(s) {
			continue
		}
		id, _ := primitive.ObjectIDFromHex(s)
		recipientIDs = append(recipientIDs, id)
	}

	// Create the message
	anon := utils.Ternary(args.Anonymous != nil && *args.Anonymous, true, false).(bool)
	important := utils.Ternary(args.Important != nil && *args.Important, true, false).(bool)
	mb := structures.NewMessageBuilder(&structures.Message{}).
		SetKind(structures.MessageKindInbox).
		SetAnonymous(anon).
		SetTimestamp(time.Now()).
		SetAuthorID(actor.ID).
		AsInbox(structures.MessageDataInbox{
			Subject:   args.Subject,
			Content:   args.Content,
			Important: important,
		})

	mm := mutations.MessageMutation{
		MessageBuilder: mb,
	}
	if _, err := mm.SendInboxMessage(ctx, r.Ctx.Inst().Mongo, mutations.SendInboxMessageOptions{
		Actor:                actor,
		Recipients:           recipientIDs,
		ConsiderBlockedUsers: !actor.HasPermission(structures.RolePermissionManageUsers),
	}); err != nil {
		return nil, err
	}

	return query.CreateMessageResolver(r.Ctx, ctx, nil, &mm.MessageBuilder.Message.ID, query.GenerateSelectedFieldMap(ctx).Children)
}
