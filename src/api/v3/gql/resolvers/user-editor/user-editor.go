package user_editor

import (
	"context"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.UserEditorResolver {
	return &Resolver{r}
}

func (r *Resolver) User(ctx context.Context, obj *model.UserEditor) (*model.User, error) {
	if obj.User != nil && obj.User.ID != structures.DeletedEmote.ID {
		return obj.User, nil
	}
	return loaders.For(ctx).UserByID.Load(obj.ID)
}