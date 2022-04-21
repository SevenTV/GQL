package loaders

import (
	"context"

	"github.com/SevenTV/Common/dataloader"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/v3/loaders"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/global"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const LoadersKey = utils.Key("dataloaders")

type Loaders struct {
	// User Loaders
	UserByID       *UserLoader
	UsersByEmoteID *loaders.BatchUserLoader
	UsersByRoleID  *loaders.BatchUserLoader

	// Emote Loaders
	EmoteByID         *EmoteLoader
	EmotesByChannelID *loaders.BatchEmoteLoader

	// Emote Set Loaders
	EmoteSetByID     *EmoteSetLoader
	EmoteSetByUserID *BatchEmoteSetLoader

	// Role Loaders
	RoleByID *loaders.RoleLoader

	// Report Loaders
	ReportByID       *loaders.ReportLoader
	ReportsByUserID  *loaders.BatchReportLoader
	ReportsByEmoteID *loaders.BatchReportLoader
}

func New(gCtx global.Context) *Loaders {
	return &Loaders{
		UserByID:         userByID(gCtx),
		EmoteByID:        emoteByID(gCtx),
		EmoteSetByID:     emoteSetByID(gCtx),
		EmoteSetByUserID: emoteSetByUserID(gCtx),
	}
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(LoadersKey).(*Loaders)
}

type (
	EmoteLoader         = dataloader.DataLoader[primitive.ObjectID, *model.Emote]
	UserLoader          = dataloader.DataLoader[primitive.ObjectID, *model.User]
	BatchUserLoader     = dataloader.DataLoader[primitive.ObjectID, []*model.User]
	EmoteSetLoader      = dataloader.DataLoader[primitive.ObjectID, *model.EmoteSet]
	BatchEmoteSetLoader = dataloader.DataLoader[primitive.ObjectID, []*model.EmoteSet]
)
