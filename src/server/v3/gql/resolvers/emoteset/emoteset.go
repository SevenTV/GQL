package emoteset

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/mutations"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/auth"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/server/v3/gql/types"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func NewOps(r types.Resolver) generated.EmoteSetOpsResolver {
	return &Resolver{r}
}

func (r *Resolver) Emotes(ctx context.Context, obj *model.EmoteSetOps, id primitive.ObjectID, action model.ListItemAction, nameArg *string) ([]*model.Emote, error) {
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

	emoteIDs := make([]primitive.ObjectID, len(b.EmoteSet.Emotes))
	for i, e := range b.EmoteSet.Emotes {
		emoteIDs[i] = e.ID
	}
	emotes, errs := loaders.For(ctx).EmoteByID.LoadAll(emoteIDs)
	return emotes, multierror.Append(nil, errs...).ErrorOrNil()
}
