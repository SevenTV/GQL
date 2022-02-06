package user_connections

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/server/v3/gql/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.UserConnectionResolver {
	return &Resolver{r}
}

func (r *Resolver) EmoteSet(ctx context.Context, obj *model.UserConnection) (*model.EmoteSet, error) {
	return loaders.For(ctx).EmoteSetByID.Load(obj.EmoteSet.ID)
}
