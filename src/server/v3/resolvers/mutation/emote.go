package mutation

import (
	"context"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/structures/mutations"
	"github.com/SevenTV/Common/utils"
	"github.com/sirupsen/logrus"

	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) EditEmote(ctx context.Context, args struct {
	EmoteID string
	Data    EditEmoteInput
}) (*query.EmoteResolver, error) {
	// Get current actor
	actor, _ := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrAccessDenied
	}

	// Parse Emote ID
	emoteID, err := primitive.ObjectIDFromHex(args.EmoteID)
	if err != nil {
		return nil, err
	}

	// Find the emote
	emote := &structures.Emote{}
	if err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmotes).FindOne(ctx, bson.M{
		"_id": emoteID,
	}).Decode(emote); err != nil {
		return nil, err
	}

	// Add the actor's edited users
	p := append(mongo.Pipeline{
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": actor}}},
	}, aggregations.UserRelationEditorOf...)
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, p)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	cur.Next(ctx)
	if err = cur.Decode(actor); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	cur.Close(ctx)

	// Add mutations
	eb := structures.NewEmoteBuilder(emote)
	if args.Data.Name != nil {
		eb.SetName(*args.Data.Name)
	}
	if args.Data.Flags != nil {
		sum := *args.Data.Flags
		hasFlag := func(flag structures.EmoteFlag) bool {
			if !utils.BitField.HasBits(int64(emote.Flags), int64(flag)) && utils.BitField.HasBits(int64(sum), int64(flag)) {
				return true
			} else if utils.BitField.HasBits(int64(emote.Flags), int64(flag)) && !utils.BitField.HasBits(int64(sum), int64(flag)) {
				return true
			}
			return false
		}

		// Check permissions for "Listed" privileged flag
		if hasFlag(structures.EmoteFlagsListed) && !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
			return nil, helpers.ErrAccessDenied
		}
		eb.SetFlags(structures.EmoteFlag(sum))
	}
	if args.Data.OwnerID != nil {
		ownerID, err := primitive.ObjectIDFromHex(*args.Data.OwnerID)
		if err != nil {
			return nil, helpers.ErrBadObjectID
		}

		eb.SetOwnerID(ownerID)
	}
	if args.Data.Tags != nil {
		eb.SetTags(*args.Data.Tags, true)
	}

	// Update the emote
	em := mutations.EmoteMutation{
		EmoteBuilder: eb,
	}
	if _, err = em.Edit(ctx, r.Ctx.Inst().Mongo, mutations.EmoteEditOptions{
		Actor: actor,
	}); err != nil {
		return nil, err
	}

	fields := query.GenerateSelectedFieldMap(ctx)
	return query.CreateEmoteResolver(r.Ctx, ctx, emote, &emote.ID, fields.Children)
}

type EditEmoteInput struct {
	Name    *string   `json:"name"`
	Flags   *int32    `json:"flags"`
	OwnerID *string   `json:"owner_id"`
	Tags    *[]string `json:"tags"`
}
