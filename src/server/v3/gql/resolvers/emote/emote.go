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
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.EmoteResolver {
	return &Resolver{r}
}

func (r *Resolver) Urls(ctx context.Context, obj *model.Emote, format *model.ImageFormat) ([]string, error) {
	result := make([]string, len(obj.Urls))
	for i, u := range obj.Urls {
		ext := ""
		if format != nil {
			switch *format {
			case model.ImageFormatWebp:
				ext = ".webp"
			case model.ImageFormatAvif:
				ext = ".avif"
			case model.ImageFormatGif:
				ext = ".gif"
			case model.ImageFormatPng:
				ext = ".png"
			}
		}
		result[i] = u + ext
	}

	return result, nil
}

func (r *Resolver) Owner(ctx context.Context, obj *model.Emote) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.Owner.ID)
}

func (r *Resolver) Channels(ctx context.Context, obj *model.Emote, limit *int, afterID string) ([]*model.User, error) {
	return loaders.For(ctx).UsersByEmoteID.Load(obj.ID.Hex())
}

func (r *Resolver) ChannelCount(ctx context.Context, obj *model.Emote) (int, error) {
	count, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).CountDocuments(ctx, bson.M{
		"channel_emotes.id": obj.ID,
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
