package mutation

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

	// Find recipients
	recipients := []*structures.User{}
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Find(ctx, bson.M{
		"$and": bson.A{
			bson.M{"_id": recipientIDs},
			bson.M{"_id": bson.M{"$not": bson.M{"$eq": actor.BlockedUserIDs}}},
		},
	})
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	if err = cur.All(ctx, &recipients); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
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

	// Write message to DB
	result, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameMessages).InsertOne(ctx, mb.Message)
	if err != nil {
		logrus.WithError(err).WithField("actor_id", actor.ID).Error("mongo, failed to create message")
		return nil, err
	}
	msgID := result.InsertedID.(primitive.ObjectID)

	// Create read states for the recipients
	w := make([]mongo.WriteModel, len(recipientIDs))
	for i, id := range recipientIDs {
		w[i] = &mongo.InsertOneModel{
			Document: &structures.MessageRead{
				MessageID:   msgID,
				RecipientID: id,
				Read:        false,
			},
		}
	}
	if _, err = r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameMessagesRead).BulkWrite(ctx, w); err != nil {
		logrus.WithError(err).WithField("message_id", result.InsertedID).Error("mongo, couldn't create a read state for message")
	}

	return query.CreateMessageResolver(r.Ctx, ctx, nil, &msgID, query.GenerateSelectedFieldMap(ctx).Children)
}
