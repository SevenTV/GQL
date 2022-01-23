package user_emote

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

func New(r types.Resolver) generated.UserEmoteResolver {
	return &Resolver{r}
}

func (r *Resolver) Emote(ctx context.Context, obj *model.UserEmote) (*model.Emote, error) {
	return loaders.For(ctx).EmoteByID.Load(obj.Emote.ID)
}
