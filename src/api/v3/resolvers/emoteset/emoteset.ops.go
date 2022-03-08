package emoteset

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
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

	// Get the emote set
	name := ""
	if nameArg != nil {
		name = *nameArg
	}
	b := structures.NewEmoteSetBuilder(nil)
	if err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmoteSets).FindOne(ctx, bson.M{
		"_id": obj.ID,
	}).Decode(b.EmoteSet); err != nil {
		logF.WithError(err).Error("mongo, couldn't find emote to add to set")
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Mutate the thing
	m := mutations.EmoteSetMutation{EmoteSetBuilder: b}
	if _, err := m.SetEmote(ctx, r.Ctx.Inst().Mongo, mutations.EmoteSetMutationSetEmoteOptions{
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

	// Publish updates for;
	// emote set, owner of emote set, actor
	go func() {
		// Find users that have this set active
		sentToActor := false
		sentToOwner := false
		cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Find(ctx, bson.M{
			"connections.emote_set_id": b.EmoteSet.ID,
		})
		if err == nil {
			for cur.Next(ctx) {
				u := &structures.User{}
				if err = cur.Decode(u); err != nil {
					continue
				}

				events.Publish(r.Ctx, "users", u.ID)
				if u.ID == actor.ID {
					sentToActor = true
				} else if u.ID == b.EmoteSet.OwnerID {
					sentToOwner = true
				}
			}
		}

		// Publish an emote set update
		events.Publish(r.Ctx, "emote_sets", b.EmoteSet.ID)
		// Send user update for set owner
		if !sentToOwner && b.EmoteSet.OwnerID != actor.ID {
			events.Publish(r.Ctx, "users", b.EmoteSet.OwnerID)
		}
		// Send user update for actor
		if !sentToActor {
			events.Publish(r.Ctx, "users", actor.ID)
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
