package user_emote

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

func New(r types.Resolver) generated.UserEmoteResolver {
	return &Resolver{r}
}

func (r *Resolver) Emote(ctx context.Context, obj *model.UserEmote) (*model.Emote, error) {
	return loaders.For(ctx).EmoteByID.Load(obj.Emote.ID)
}