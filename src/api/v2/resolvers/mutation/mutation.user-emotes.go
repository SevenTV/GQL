package mutation

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/mutations"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/events"
	"github.com/SevenTV/GQL/src/api/v2/loaders"
	"github.com/SevenTV/GQL/src/api/v3/auth"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) AddChannelEmote(ctx context.Context, channelIDArg, emoteIDArg string, reasonArg *string) (*model.User, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, errors.ErrUnauthorized()
	}

	// Parse passed arguments
	channelID, er1 := primitive.ObjectIDFromHex(channelIDArg)
	emoteID, er2 := primitive.ObjectIDFromHex(emoteIDArg)
	if err := multierror.Append(er1, er2).ErrorOrNil(); err != nil {
		return nil, errors.ErrBadObjectID()
	}

	// Get the target user
	target, err := loaders.For(ctx).UserByID.Load(channelID.Hex())
	if err != nil {
		return nil, err
	}

	// Get the emote set
	setID, _ := primitive.ObjectIDFromHex(target.EmoteSetID)
	if setID.IsZero() {
		return nil, errors.ErrUnknownEmoteSet()
	}
	b := structures.NewEmoteSetBuilder(nil)
	if err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmoteSets).FindOne(ctx, bson.M{
		"_id": setID,
	}).Decode(b.EmoteSet); err != nil {
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Run mutation
	if err = r.doSetChannelEmote(ctx, actor, emoteID, "", mutations.ListItemActionAdd, b); err != nil {
		return nil, err
	}

	return loaders.For(ctx).UserByID.Load(channelID.Hex())
}

func (r *Resolver) RemoveChannelEmote(ctx context.Context, channelIDArg, emoteIDArg string, reasonArg *string) (*model.User, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, errors.ErrUnauthorized()
	}

	// Parse passed arguments
	channelID, er1 := primitive.ObjectIDFromHex(channelIDArg)
	emoteID, er2 := primitive.ObjectIDFromHex(emoteIDArg)
	if err := multierror.Append(er1, er2).ErrorOrNil(); err != nil {
		return nil, errors.ErrBadObjectID()
	}

	// Get the target user
	target, err := loaders.For(ctx).UserByID.Load(channelID.Hex())
	if err != nil {
		return nil, err
	}

	// Get the emote set
	setID, _ := primitive.ObjectIDFromHex(target.EmoteSetID)
	if setID.IsZero() {
		return nil, errors.ErrUnknownEmoteSet()
	}
	b := structures.NewEmoteSetBuilder(nil)
	if err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmoteSets).FindOne(ctx, bson.M{
		"_id": setID,
	}).Decode(b.EmoteSet); err != nil {
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Run mutation
	if err = r.doSetChannelEmote(ctx, actor, emoteID, "", mutations.ListItemActionRemove, b); err != nil {
		return nil, err
	}

	return loaders.For(ctx).UserByID.Load(channelID.Hex())
}

func (r *Resolver) EditChannelEmote(ctx context.Context, channelIDArg string, emoteIDArg string, data model.ChannelEmoteInput, reason *string) (*model.User, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, errors.ErrUnauthorized()
	}

	// Data must contain alias
	alias := ""
	if data.Alias != nil {
		alias = *data.Alias
	}

	// Parse passed arguments
	channelID, er1 := primitive.ObjectIDFromHex(channelIDArg)
	emoteID, er2 := primitive.ObjectIDFromHex(emoteIDArg)
	if err := multierror.Append(er1, er2).ErrorOrNil(); err != nil {
		return nil, errors.ErrBadObjectID()
	}

	// Get the target user
	target, err := loaders.For(ctx).UserByID.Load(channelID.Hex())
	if err != nil {
		return nil, err
	}

	// Get the emote set
	setID, _ := primitive.ObjectIDFromHex(target.EmoteSetID)
	if setID.IsZero() {
		return nil, errors.ErrUnknownEmoteSet()
	}
	b := structures.NewEmoteSetBuilder(nil)
	if err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmoteSets).FindOne(ctx, bson.M{
		"_id": setID,
	}).Decode(b.EmoteSet); err != nil {
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Run mutation
	if err = r.doSetChannelEmote(ctx, actor, emoteID, alias, mutations.ListItemActionUpdate, b); err != nil {
		return nil, err
	}

	return loaders.For(ctx).UserByID.Load(channelID.Hex())
}

func (r *Resolver) doSetChannelEmote(
	ctx context.Context,
	actor *structures.User,
	emoteID primitive.ObjectID,
	name string,
	action mutations.ListItemAction,
	b *structures.EmoteSetBuilder,
) error {
	m := mutations.EmoteSetMutation{EmoteSetBuilder: b}
	if _, err := m.SetEmote(ctx, r.Ctx.Inst().Mongo, mutations.EmoteSetMutationSetEmoteOptions{
		Actor: actor,
		Emotes: []mutations.EmoteSetMutationSetEmoteItem{{
			Action: action,
			ID:     emoteID,
			Name:   name,
		}},
	}); err != nil {
		logrus.WithError(err).Error("failed to update emotes in set")
		return err
	}

	// Publish an emote set update
	go func() {
		events.Publish(r.Ctx, "emote_sets", b.EmoteSet.ID)
	}()
	return nil
}