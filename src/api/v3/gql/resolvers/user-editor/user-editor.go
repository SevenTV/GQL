package user_editor

import (
	"context"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.UserEditorResolver {
	return &Resolver{r}
}

func (r *Resolver) User(ctx context.Context, obj *model.UserEditor) (*model.UserPartial, error) {
	if obj.User != nil && obj.User.ID != structures.DeletedEmote.ID {
		return obj.User, nil
	}
	u, err := loaders.For(ctx).UserByID.Load(obj.ID)
	if err != nil {
		return nil, err
	}
	return helpers.UserStructureToPartialModel(r.Ctx, u), nil
}
