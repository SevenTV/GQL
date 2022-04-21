package emoteset

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/mutations"
	"github.com/SevenTV/GQL/graph/v3/generated"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/events"
	"github.com/SevenTV/GQL/src/api/v3/auth"
	"github.com/SevenTV/GQL/src/api/v3/helpers"
	"github.com/SevenTV/GQL/src/api/v3/loaders"
	"github.com/SevenTV/GQL/src/api/v3/types"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResolverOps struct {
	types.Resolver
}

func NewOps(r types.Resolver) generated.EmoteSetOpsResolver {
	return &ResolverOps{r}
}

func (r *ResolverOps) Emotes(ctx context.Context, obj *model.EmoteSetOps, id primitive.ObjectID, action model.ListItemAction, nameArg *string) ([]*model.ActiveEmote, error) {
	actor := auth.For(ctx)
	logF := logrus.WithFields(logrus.Fields{
		"emote_set_id": obj.ID,
		"emote_id":     id,
	})

	// Get the emote
	emote, err := r.Ctx.Inst().Query.Emotes(ctx, bson.M{"_id": id}).First()
	if err != nil {
		if errors.Compare(err, errors.ErrNoItems()) {
			return nil, errors.ErrUnknownEmote()
		}
		return nil, err
	}

	// Get the emote set
	name := ""
	if nameArg != nil {
		name = *nameArg
	}
	set, err := r.Ctx.Inst().Query.EmoteSets(ctx, bson.M{"_id": obj.ID}).First()
	if err != nil {
		if errors.Compare(err, errors.ErrNoItems()) {
			return nil, errors.ErrUnknownEmoteSet()
		}
		return nil, err
	}
	b := structures.NewEmoteSetBuilder(set)

	// Mutate the thing
	if err := r.Ctx.Inst().Mutate.EditEmotesInSet(ctx, b, mutations.EmoteSetMutationSetEmoteOptions{
		Actor: actor,
		Emotes: []mutations.EmoteSetMutationSetEmoteItem{{
			Action: mutations.ListItemAction(action),
			ID:     id,
			Name:   name,
			Flags:  0,
		}},
	}); err != nil {
		logF.WithError(err).Error("failed to update emotes in set")
		return nil, err
	}

	// Clear cache keys for active sets / channel count
	k := r.Ctx.Inst().Redis.ComposeKey("gql-v3", fmt.Sprintf("emote:%s", id.Hex()))
	_, _ = r.Ctx.Inst().Redis.Del(ctx, k+":active_sets")
	_, _ = r.Ctx.Inst().Redis.Del(ctx, k+":channel_count")

	emoteIDs := make([]primitive.ObjectID, len(b.EmoteSet.Emotes))
	for i, e := range b.EmoteSet.Emotes {
		emoteIDs[i] = e.ID
	}

	// Publish an emote set update
	go func() {
		events.Publish(r.Ctx, "emote_sets", b.EmoteSet.ID)

		// Legacy Event API v1
		if set.Owner != nil {
			tw, _, err := set.Owner.Connections.Twitch()
			if err != nil {
				return
			}
			if tw.EmoteSetID.IsZero() || tw.EmoteSetID != set.ID {
				return // skip if target emote set isn't bound to user connection
			}
			events.PublishLegacyEventAPI(r.Ctx, action.String(), actor, set, emote, tw.Data.Login)
		}
	}()

	setModel := helpers.EmoteSetStructureToModel(r.Ctx, b.EmoteSet)
	emotes, errs := loaders.For(ctx).EmoteByID.LoadAll(emoteIDs)
	for i, e := range emotes {
		if ae := setModel.Emotes[i]; ae != nil {
			setModel.Emotes[i].Emote = e
		}
	}

	return setModel.Emotes, multierror.Append(nil, errs...).ErrorOrNil()
}
