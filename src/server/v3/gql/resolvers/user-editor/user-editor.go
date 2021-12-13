package user_editor

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

func New(r types.Resolver) generated.UserEditorResolver {
	return &Resolver{r}
}

func (r *Resolver) User(ctx context.Context, obj *model.UserEditor) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.User.ID)
}
