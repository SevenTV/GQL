package subscription

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/auth"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) CurrentUser(ctx context.Context, init *bool) (<-chan *model.UserPartial, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, nil
	}

	getUser := func() (*model.UserPartial, error) {
		user, err := loaders.For(ctx).UserByID.Load(actor.ID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.ErrUnknownUser()
			}

			logrus.WithError(err).Error("failed to subscribe")
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}
		return helpers.UserStructureToPartialModel(r.Ctx, user), nil
	}

	ch := make(chan *model.UserPartial, 1)
	if init != nil && *init {
		user, err := getUser()
		if err != nil {
			return nil, err
		}
		ch <- user
	}

	go func() {
		defer close(ch)
		sub := r.subscribe(ctx, "users", actor.ID)
		for range sub {
			user, _ := getUser()
			if user != nil {
				ch <- user
			}
		}
	}()

	return ch, nil
}

func (r *Resolver) User(ctx context.Context, id primitive.ObjectID, init *bool) (<-chan *model.UserPartial, error) {
	getUser := func() (*model.UserPartial, error) {
		user, err := loaders.For(ctx).UserByID.Load(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.ErrUnknownUser()
			}

			logrus.WithError(err).Error("failed to subscribe")
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}
		return helpers.UserStructureToPartialModel(r.Ctx, user), nil
	}

	ch := make(chan *model.UserPartial, 1)
	if init != nil && *init {
		user, err := getUser()
		if err != nil {
			return nil, err
		}
		ch <- user
	}

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
