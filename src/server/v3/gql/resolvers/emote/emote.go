package emote

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/server/v3/gql/types"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.EmoteResolver {
	return &Resolver{r}
}

func (r *Resolver) Owner(ctx context.Context, obj *model.Emote) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.Owner.ID)
}

func (r *Resolver) Channels(ctx context.Context, obj *model.Emote, limit *int, afterID string) ([]*model.User, error) {
	return loaders.For(ctx).UsersByEmoteID.Load(obj.ID)
}

func (r *Resolver) ChannelCount(ctx context.Context, obj *model.Emote) (int, error) {
	id, err := primitive.ObjectIDFromHex(obj.ID)
	if err != nil {
		return 0, nil
	}

	count, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).CountDocuments(ctx, bson.M{
		"channel_emotes.id": id,
	})
	if err != nil {
		logrus.WithError(err).Error("failed to count documents for emotes")
		return 0, err
	}

	return int(count), nil
}

func (r *Resolver) Reports(ctx context.Context, obj *model.Emote) ([]*model.Report, error) {
	return loaders.For(ctx).ReportsByEmoteID.Load(obj.ID)
}
