package helpers

import (
	"strconv"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/sirupsen/logrus"
)

// UserStructureToModel: Transform a user structure to a GQL mdoel
func UserStructureToModel(s *structures.User) *model.User {
	if s == nil {
		return UserStructureToModel(structures.DeletedUser)
	}
	tagColor := 0
	if role := s.GetHighestRole(); role != nil {
		tagColor = int(role.Color)
	}
	roles := make([]*model.Role, len(s.Roles))
	for i, v := range s.Roles {
		roles[i] = RoleStructureToModel(v)
	}

	connections := make([]*model.UserConnection, len(s.Connections))
	for i, v := range s.Connections {
		connections[i] = UserConnectionStructureToModel(v)
	}

	editors := make([]*model.UserEditor, len(s.Editors))
	for i, v := range s.Editors {
		editors[i] = UserEditorStructureToModel(v)
	}

	return &model.User{
		ID:               s.ID,
		UserType:         string(s.UserType),
		Username:         s.Username,
		DisplayName:      utils.Ternary(len(s.DisplayName) > 0, s.DisplayName, s.Username).(string),
		CreatedAt:        s.ID.Timestamp(),
		AvatarURL:        "",
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
func UserEditorStructureToModel(s *structures.UserEditor) *model.UserEditor {
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
		User:        UserStructureToModel(s.User),
	}
}

// UserConnectionStructureToModel: Transform a user connection structure to a GQL model
func UserConnectionStructureToModel(s *structures.UserConnection) *model.UserConnection {
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
	}
}

// RoleStructureToModel: Transform a role structure to a GQL model
func RoleStructureToModel(s *structures.Role) *model.Role {
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

func EmoteStructureToModel(s *structures.Emote) *model.Emote {
	return &model.Emote{
		ID:           s.ID,
		Name:         s.Name,
		Flags:        int(s.Flags),
		Status:       int(s.Status),
		Tags:         s.Tags,
		Animated:     s.FrameCount > 1,
		CreatedAt:    s.ID.Timestamp(),
		Owner:        UserStructureToModel(s.Owner),
		Channels:     []*model.User{},
		ChannelCount: 0,
		Links:        [][]string{},
		Reports:      []*model.Report{},
	}
}
