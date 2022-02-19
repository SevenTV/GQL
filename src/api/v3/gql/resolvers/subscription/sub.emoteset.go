package subscription

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) EmoteSet(ctx context.Context, id primitive.ObjectID, init *bool) (<-chan *model.EmoteSet, error) {
	getEmoteSet := func() (*model.EmoteSet, error) {
		set, err := loaders.For(ctx).EmoteSetByID.Load(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.ErrUnknownEmote()
			}

			logrus.WithError(err).Error("failed to subscribe")
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}
		return set, nil
	}

	ch := make(chan *model.EmoteSet, 1)
	if init != nil && *init {
		set, err := getEmoteSet()
		if err != nil {
			return nil, err
		}
		ch <- set
	}

	go func() {
		defer close(ch)
		sub := r.subscribe(ctx, "emote_sets", id)
		for range sub {
			set, _ := getEmoteSet()
			if set != nil {
				ch <- set
			}
		}
	}()

	return ch, nil
}
