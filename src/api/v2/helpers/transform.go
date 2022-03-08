package helpers

import (
	"time"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/global"
)

func EmoteStructureToModel(ctx global.Context, s *structures.Emote) *model.Emote {
	eb := structures.NewEmoteBuilder(s)
	version, _ := eb.GetVersion(s.ID)

	return &model.Emote{
		ID:           s.ID.Hex(),
		Name:         s.Name,
		OwnerID:      s.OwnerID.Hex(),
		Visibility:   0, // TODO
		Mime:         "image/webp",
		Status:       int(version.State.Lifecycle),
		Tags:         s.Tags,
		CreatedAt:    s.ID.Timestamp().Format(time.RFC3339),
		AuditEntries: []*model.AuditLog{},
		Channels:     []*model.UserPartial{},
		ChannelCount: int(version.State.ChannelCount),
		Owner:        nil,
		Reports:      []*model.Report{},
		Urls:         [][]string{},
		Width:        []int{},
		Height:       []int{},
	}
}
