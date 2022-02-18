package subscription

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) Emote(ctx context.Context, id primitive.ObjectID, init *bool) (<-chan *model.EmotePartial, error) {
	getEmote := func() (*model.EmotePartial, error) {
		emote, err := loaders.For(ctx).EmoteByID.Load(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.ErrUnknownEmote()
			}

			logrus.WithError(err).Error("failed to subscribe")
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}
		return helpers.EmoteStructureToPartialModel(r.Ctx, emote), nil
	}

	emote, err := getEmote()
	if err != nil {
		return nil, err
	}
	ch := make(chan *model.EmotePartial, 1)
	ch <- emote

	go func() {
		sub := r.subscribe(ctx, "emotes", id)
		for range sub {
			emote, _ := getEmote()
			if emote != nil {
				ch <- emote
			}
		}
	}()

	return ch, nil
}
