package role

import (
	"context"

	"github.com/SevenTV/GQL/graph/v3/generated"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.RoleResolver {
	return &Resolver{r}
}

func (r *Resolver) Members(ctx context.Context, obj *model.Role, page *int, limit *int) ([]*model.User, error) {
	// TODO
	return nil, nil
}
