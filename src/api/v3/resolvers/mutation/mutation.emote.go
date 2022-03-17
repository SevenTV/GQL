package mutation

import (
	"context"

	"github.com/SevenTV/GQL/graph/v3/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (*Resolver) Emote(ctx context.Context, id primitive.ObjectID) (*model.EmoteOps, error) {
	return &model.EmoteOps{
		ID: id,
	}, nil
}
