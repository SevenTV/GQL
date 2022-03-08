package loaders

import (
	"context"

	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/v3/loaders"
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

	// Emote Set Loaders
	EmoteSetByID     *loaders.EmoteSetLoader
	EmoteSetByUserID *loaders.BatchEmoteSetLoader

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
