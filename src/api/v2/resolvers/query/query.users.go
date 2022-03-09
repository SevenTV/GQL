package query

import (
	"context"

	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/loaders"
	"github.com/SevenTV/GQL/src/api/v3/auth"
)

func (r *Resolver) User(ctx context.Context, id string) (*model.User, error) {
	// Handle @me (fetch actor)
	// this sets the queried user ID to that of the actor user
	if id == "@me" {
		actor := auth.For(ctx)
		id = actor.ID.Hex()
	}

	return loaders.For(ctx).UserByID.Load(id)
}
