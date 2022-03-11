package query

import (
	"context"
	"strings"

	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/loaders"
	"github.com/SevenTV/GQL/src/api/v3/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) User(ctx context.Context, id string) (*model.User, error) {
	if primitive.IsValidObjectID(id) {
		return loaders.For(ctx).UserByID.Load(id)
	} else if id == "@me" {
		// Handle @me (fetch actor)
		// this sets the queried user ID to that of the actor user
		actor := auth.For(ctx)
		id = actor.ID.Hex()
		return loaders.For(ctx).UserByID.Load(id)
	} else {
		// at this point we assume the query is for a username
		// (it was neither an id, or the @me label)
		return loaders.For(ctx).UserByUsername.Load(strings.ToLower(id))
	}
}
