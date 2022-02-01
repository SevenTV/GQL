package helpers

import (
	"fmt"
	"strconv"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/global"
	"github.com/sirupsen/logrus"
)

// UserStructureToModel: Transform a user structure to a GQL mdoel
func UserStructureToModel(ctx global.Context, s *structures.User) *model.User {
	tagColor := 0
	if role := s.GetHighestRole(); role != nil {
		tagColor = int(role.Color)
	}
	roles := make([]*model.Role, len(s.Roles))
	for i, v := range s.Roles {
		roles[i] = RoleStructureToModel(ctx, v)
	}

	connections := make([]*model.UserConnection, len(s.Connections))
	for i, v := range s.Connections {
		connections[i] = UserConnectionStructureToModel(ctx, v)
	}

	editors := make([]*model.UserEditor, len(s.Editors))
	for i, v := range s.Editors {
		editors[i] = UserEditorStructureToModel(ctx, v)
	}

	avatarURL := ""
	if s.AvatarID != "" {
		avatarURL = fmt.Sprintf("//%s/pp/%s/%s", ctx.Config().CdnURL, s.ID.Hex(), s.AvatarID)
	} else {
		for _, con := range s.Connections {
			switch con.Platform {
			case structures.UserConnectionPlatformTwitch:
				if d, err := con.DecodeTwitch(); err == nil {
					avatarURL = d.ProfileImageURL[6:]
				}
			}
		}
	}

	return &model.User{
		ID:               s.ID,
		UserType:         string(s.UserType),
		Username:         s.Username,
		DisplayName:      utils.Ternary(len(s.DisplayName) > 0, s.DisplayName, s.Username).(string),
		CreatedAt:        s.ID.Timestamp(),
		AvatarURL:        avatarURL,
		Biography:        s.Biography,
		TagColor:         tagColor,
		Editors:          editors,
		Roles:            roles,
		OwnedEmotes:      []*model.Emote{},
		Connections:      connections,
		InboxUnreadCount: 0,
		Reports:          []*model.Report{},
	}
}

// UserEditorStructureToModel: Transform a user editor structure to a GQL model
func UserEditorStructureToModel(ctx global.Context, s *structures.UserEditor) *model.UserEditor {
	if s.User == nil {
		s.User = structures.DeletedUser
	}
	connIDs := make([]string, len(s.Connections))
	for i, connID := range s.Connections {
		connIDs[i] = connID.Hex()
	}

	return &model.UserEditor{
		Connections: connIDs,
		Permissions: int(s.Permissions),
		Visible:     s.Visible,
		AddedAt:     s.AddedAt,
		User:        UserStructureToModel(ctx, s.User),
	}
}

// UserConnectionStructureToModel: Transform a user connection structure to a GQL model
func UserConnectionStructureToModel(ctx global.Context, s *structures.UserConnection) *model.UserConnection {
	var (
		err         error
		d           interface{}
		displayName string
	)
	// Decode the connection data
	switch s.Platform {
	case structures.UserConnectionPlatformTwitch:
		if d, err = s.DecodeTwitch(); err == nil {
			displayName = d.(*structures.TwitchConnection).DisplayName
		}
	case structures.UserConnectionPlatformYouTube:
		if d, err = s.DecodeYouTube(); err == nil {
			displayName = d.(*structures.YouTubeConnection).Title
		}
	}
	if err != nil {
		logrus.WithError(err).Errorf("couldn't decode %s user connection", s.Platform)
	}

	return &model.UserConnection{
		ID:          s.ID,
		DisplayName: displayName,
		Platform:    string(s.Platform),
		LinkedAt:    s.LinkedAt,
		EmoteSetID:  &s.EmoteSetID,
	}
}

// RoleStructureToModel: Transform a role structure to a GQL model
func RoleStructureToModel(ctx global.Context, s *structures.Role) *model.Role {
	return &model.Role{
		ID:        s.ID,
		Name:      s.Name,
		Color:     int(s.Color),
		Allowed:   strconv.Itoa(int(s.Allowed)),
		Denied:    strconv.Itoa(int(s.Denied)),
		Position:  int(s.Position),
		CreatedAt: s.ID.Timestamp(),
		Members:   []*model.User{},
	}
}

func EmoteStructureToModel(ctx global.Context, s *structures.Emote) *model.Emote {
	urls := make([]string, 4)
	for i := range urls {
		size := strconv.Itoa(i + 1)
		urls[i] = fmt.Sprintf("//%s/emote/%s/%sx", ctx.Config().CdnURL, s.ID.Hex(), size)
	}

	owner := utils.Ternary(s.Owner != nil, s.Owner, structures.DeletedUser).(*structures.User)
	return &model.Emote{
		ID:           s.ID,
		Name:         s.Name,
		Flags:        int(s.Flags),
		Status:       int(s.Status),
		Tags:         s.Tags,
		Animated:     s.FrameCount > 1,
		CreatedAt:    s.ID.Timestamp(),
		Owner:        UserStructureToModel(ctx, owner),
		Channels:     []*model.User{},
		ChannelCount: 0,
		Urls:         urls,
		Reports:      []*model.Report{},
	}
}

func EmoteSetStructureToModel(ctx global.Context, s *structures.EmoteSet) *model.EmoteSet {
	emotes := make([]*model.ActiveEmote, len(s.Emotes))
	for i, e := range s.Emotes {
		emotes[i] = &model.ActiveEmote{
			ID:        e.ID,
			Name:      e.Name,
			Flags:     int(e.Flags),
			Timestamp: e.Timestamp,
		}
	}

	return &model.EmoteSet{
		ID:         s.ID,
		Name:       s.Name,
		Tags:       s.Tags,
		Emotes:     emotes,
		EmoteSlots: int(s.EmoteSlots),
		Owner:      UserStructureToModel(ctx, s.Owner),
	}
}
