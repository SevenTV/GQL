package subscription

import (
	"context"

	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/auth"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) CurrentUser(ctx context.Context, init *bool) (<-chan *model.UserPartial, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, nil
	}

	getUser := func() *model.UserPartial {
		user, err := loaders.For(ctx).UserByID.Load(actor.ID)
		if err != nil {
			return nil
		}
		return helpers.UserStructureToPartialModel(r.Ctx, user)
	}

	ch := make(chan *model.UserPartial, 1)
	if init != nil && *init {
		user := getUser()
		if user != nil {
			ch <- user
		}
	}

	go func() {
		defer close(ch)
		sub := r.subscribe(ctx, "users", actor.ID)
		for range sub {
			user := getUser()
			if user != nil {
				ch <- user
			}
		}
	}()

	return ch, nil
}

func (r *Resolver) User(ctx context.Context, id primitive.ObjectID, init *bool) (<-chan *model.UserPartial, error) {
	getUser := func() *model.UserPartial {
		user, err := loaders.For(ctx).UserByID.Load(id)
		if err != nil {
			return nil
		}
		return helpers.UserStructureToPartialModel(r.Ctx, user)
	}

	ch := make(chan *model.UserPartial, 1)
	if init != nil && *init {
		user := getUser()
		if user != nil {
			ch <- user
		}
	}

	go func() {
		defer close(ch)
		sub := r.subscribe(ctx, "users", id)
		for range sub {
			user := getUser()
			if user != nil {
				ch <- user
			}
		}
	}()

	return ch, nil
}
