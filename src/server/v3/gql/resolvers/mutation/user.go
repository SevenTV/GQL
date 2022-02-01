package mutation

import (
	"context"

	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) User(ctx context.Context, id primitive.ObjectID) (*model.UserOps, error) {
	user, err := loaders.For(ctx).UserByID.Load(id)
	if err != nil {
		return nil, err
	}

	return &model.UserOps{
		ID:          user.ID,
		Connections: user.Connections,
	}, nil
}
