package query

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) EmoteSet(ctx context.Context, id primitive.ObjectID) (*model.EmoteSet, error) {
	return loaders.For(ctx).EmoteSetByID.Load(id)
}

func (r *Resolver) NamedEmoteSet(ctx context.Context, name model.EmoteSetName) (*model.EmoteSet, error) {
	var setID primitive.ObjectID

	switch name {
	case model.EmoteSetNameGlobal:
		sys := r.Ctx.Inst().Mongo.System(ctx)
		setID = sys.EmoteSetID
	}

	if setID.IsZero() {
		return nil, errors.ErrUnknownEmoteSet()
	}

	return loaders.For(ctx).EmoteSetByID.Load(setID)
}
