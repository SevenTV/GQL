package emoteset

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/loaders"
	"github.com/SevenTV/GQL/src/api/v3/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.EmoteSetResolver {
	return &Resolver{r}
}

func (r *Resolver) Owner(ctx context.Context, obj *model.EmoteSet) (*model.User, error) {
	if obj.OwnerID == nil {
		return nil, nil
	}

	return loaders.For(ctx).UserByID.Load(*obj.OwnerID)
}
