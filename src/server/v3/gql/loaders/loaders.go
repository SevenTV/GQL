package loaders

import (
	"context"

	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/loaders"
	"github.com/SevenTV/GQL/src/global"
)

const LoadersKey = utils.Key("dataloaders")

type Loaders struct {
	// User Loaders
	UserByID       *loaders.UserLoader
	UsersByEmoteID *loaders.BatchUserLoader
	UsersByRoleID  *loaders.BatchUserLoader

	// Emote Loaders
	EmoteByID         *loaders.EmoteLoader
	EmotesByChannelID *loaders.BatchEmoteLoader

	// Role Loaders
	RoleByID *loaders.RoleLoader

	// Connection Loaders
	ConnectionByID *loaders.ConnectionLoader

	// Report Loaders
	ReportByID       *loaders.ReportLoader
	ReportsByUserID  *loaders.BatchReportLoader
	ReportsByEmoteID *loaders.BatchReportLoader
}

func New(gCtx global.Context) *Loaders {
	// TODO
	return &Loaders{
		UserByID: userLoader,
	}
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(LoadersKey).(*Loaders)
}
