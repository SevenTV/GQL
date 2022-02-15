package role

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

func New(r types.Resolver) generated.RoleResolver {
	return &Resolver{r}
}

func (r *Resolver) Members(ctx context.Context, obj *model.Role) ([]*model.User, error) {
	return loaders.For(ctx).UsersByRoleID.Load(obj.ID.Hex())
}
