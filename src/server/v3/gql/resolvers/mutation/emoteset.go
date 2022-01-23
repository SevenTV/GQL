package mutation

import (
	"context"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/mutations"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/auth"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) EmoteSet(ctx context.Context, id primitive.ObjectID) (*model.EmoteSetOps, error) {
	return nil, nil
}

// CreateEmoteSet: create a new emote set
func (r *Resolver) CreateEmoteSet(ctx context.Context, input model.CreateEmoteSetInput) (*model.EmoteSet, error) {
	actor := auth.For(ctx)

	// Set up emote set builder
	isPrivileged := utils.Ternary(input.Privileged != nil, input.Privileged, false).(bool)
	b := structures.NewEmoteSetBuilder(nil).
		SetName(input.Name).
		SetPrivileged(isPrivileged).
		SetOwnerID(actor.ID).
		SetEmoteSlots(250).
		SetActive(true)
	m := mutations.EmoteSetMutation{
		EmoteSetBuilder: b,
	}

	// Execute mutation
	if _, err := m.Create(ctx, r.Ctx.Inst().Mongo, mutations.EmoteSetMutationOptions{
		Actor: actor,
	}); err != nil {
		return nil, err
	}

	return loaders.For(ctx).EmoteSetByID.Load(b.EmoteSet.ID)
}
