package ban

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.BanResolver {
	return &Resolver{r}
}

func (r *Resolver) Victim(ctx context.Context, obj *model.Ban) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.Victim.ID)
}

func (r *Resolver) Actor(ctx context.Context, obj *model.Ban) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.Actor.ID)
}
