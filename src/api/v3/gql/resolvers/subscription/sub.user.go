package subscription

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/auth"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (r *Resolver) CurrentUser(ctx context.Context) (<-chan *model.User, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, nil
	}

	getUser := func() (*model.User, error) {
		user, err := loaders.For(ctx).UserByID.Load(actor.ID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.ErrUnknownUser()
			}

			logrus.WithError(err).Error("failed to subscribe")
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}
		return user, nil
	}

	user, err := getUser()
	if err != nil {
		return nil, err
	}
	ch := make(chan *model.User, 1)
	ch <- user

	go func() {
		sub := r.subscribe(ctx, "users", actor.ID)
		for range sub {
			user, _ = getUser()
			if user != nil {
				ch <- user
			}
		}
	}()

	return ch, nil
}

func (r *Resolver) User(ctx context.Context, id primitive.ObjectID) (<-chan *model.User, error) {
	getUser := func() (*model.User, error) {
		user, err := loaders.For(ctx).UserByID.Load(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.ErrUnknownEmote()
			}

			logrus.WithError(err).Error("failed to subscribe")
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}
		return user, nil
	}

	user, err := getUser()
	if err != nil {
		return nil, err
	}
	ch := make(chan *model.User, 1)
	ch <- user

	go func() {
		sub := r.subscribe(ctx, "users", id)
		for range sub {
			user, _ := getUser()
			if user != nil {
				ch <- user
			}
		}
	}()

	return ch, nil
}
