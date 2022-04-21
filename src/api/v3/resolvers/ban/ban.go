package ban

import (
	"context"

	"github.com/SevenTV/GQL/graph/v3/generated"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/loaders"
	"github.com/SevenTV/GQL/src/api/v3/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.BanResolver {
	return &Resolver{r}
}

func (r *Resolver) Victim(ctx context.Context, obj *model.Ban) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.VictimID)
}

func (r *Resolver) Actor(ctx context.Context, obj *model.Ban) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.ActorID)
}
