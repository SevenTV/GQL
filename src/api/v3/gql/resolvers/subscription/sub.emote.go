package subscription

import (
	"context"

	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) Emote(ctx context.Context, id primitive.ObjectID, init *bool) (<-chan *model.EmotePartial, error) {
	getEmote := func() *model.EmotePartial {
		emote, err := loaders.For(ctx).EmoteByID.Load(id)
		if err != nil {
			return nil
		}
		return helpers.EmoteStructureToPartialModel(r.Ctx, emote)
	}

	ch := make(chan *model.EmotePartial, 1)
	if init != nil && *init {
		emote := getEmote()
		if emote != nil {
			ch <- emote
		}
	}

	go func() {
		defer close(ch)
		sub := r.subscribe(ctx, "emotes", id)
		for range sub {
			emote := getEmote()
			if emote != nil {
				ch <- emote
			}
		}
	}()

	return ch, nil
}
