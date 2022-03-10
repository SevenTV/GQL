package loaders

import (
	"context"

	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/v2/loaders"
	"github.com/SevenTV/GQL/src/global"
)

const LoadersKey = utils.Key("dataloadersv2")

type Loaders struct {
	UserByID   *loaders.UserLoader
	UserEmotes *loaders.UserEmotesLoader

	EmoteByID *loaders.EmoteLoader
}

func New(gCtx global.Context) *Loaders {
	return &Loaders{
		UserByID:   userByID(gCtx),
		UserEmotes: userEmotesLoader(gCtx),

		EmoteByID: emoteByID(gCtx),
	}
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(LoadersKey).(*Loaders)
}