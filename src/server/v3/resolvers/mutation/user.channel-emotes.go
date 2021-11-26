package mutation

import (
	"context"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/structures/mutations"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) SetChannelEmote(ctx context.Context, args struct {
	UserID string
	Target channelEmoteInput
	Action mutations.ListItemAction
}) (*query.UserResolver, error) {
	// Get the actor user
	actor := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrUnauthorized
	}

	// Get the target user (i.e may be a user that the actor is an editor of)
	targetUser := actor
	targetUserID, err := primitive.ObjectIDFromHex(args.UserID)
	if err != nil {
		return nil, helpers.ErrBadObjectID
	}

	// Fetch the target user if they are not the actor
	if targetUserID != targetUser.ID {
		cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, append(mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"_id": targetUserID}}},
		}, aggregations.UserRelationRoles...))
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, helpers.ErrUnknownUser
			}

			logrus.WithError(err).Error("mongo")
			return nil, helpers.ErrInternalServerError
		}

		cur.Next(ctx)
		err = cur.Decode(targetUser)
		if err != nil {
			return nil, err
		}
		err = cur.Close(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Prepare the mutation
	um := &mutations.UserMutation{
		UserBuilder: structures.NewUserBuilder(targetUser),
	}
	emoteID, err := primitive.ObjectIDFromHex(args.Target.ID)
	if err != nil {
		return nil, helpers.ErrBadObjectID
	}
	alias := ""
	if args.Target.Alias != nil { // Set alias
		alias = *args.Target.Alias
	}

	// Do the mutation
	_, err = um.SetChannelEmote(ctx, r.Ctx.Inst().Mongo, mutations.SetChannelEmoteOptions{
		Actor:    actor,
		EmoteID:  emoteID,
		Channels: []primitive.ObjectID{},
		Alias:    alias,
		Action:   args.Action,
	})
	if err != nil {
		logrus.WithError(err).Error("SetChannelEmote")
		return nil, err
	}

	return query.CreateUserResolver(r.Ctx, ctx, targetUser, &targetUser.ID, query.GenerateSelectedFieldMap(ctx).Children)
}

type channelEmoteInput struct {
	ID       string
	Channels *[]string
	Alias    *string
}
