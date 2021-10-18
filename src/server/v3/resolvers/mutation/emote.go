package mutation

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"

	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) Emote(ctx context.Context, args struct {
	ID string
}) (*query.EmoteResolver, error) {
	// Get current user
	user := ctx.Value("user").(*structures.User)
	if user == nil {
		return nil, helpers.ErrAccessDenied
	}

	// Parse Emote ID
	emoteID, err := primitive.ObjectIDFromHex(args.ID)
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

	// Verify permissions

	fields := query.GenerateSelectedFieldMap(ctx)
	return query.CreateEmoteResolver(r.Ctx, ctx, emote, &emote.ID, fields.Children)
}
