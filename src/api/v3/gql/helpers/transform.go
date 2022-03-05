package helpers

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/global"
	"github.com/sirupsen/logrus"
)

var twitchPictureSizeRegExp = regexp.MustCompile("([0-9]{2,3})x([0-9]{2,3})")

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
					avatarURL = twitchPictureSizeRegExp.ReplaceAllString(d.ProfileImageURL[6:], "70x70")
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

func UserStructureToPartialModel(ctx global.Context, m *model.User) *model.UserPartial {
	return &model.UserPartial{
		ID:          m.ID,
		UserType:    m.UserType,
		Username:    m.Username,
		DisplayName: m.DisplayName,
		CreatedAt:   m.ID.Timestamp(),
		AvatarURL:   m.AvatarURL,
		Biography:   m.Biography,
		TagColor:    m.TagColor,
		Roles:       m.Roles,
		Connections: m.Connections,
	}
}

// UserEditorStructureToModel: Transform a user editor structure to a GQL model
func UserEditorStructureToModel(ctx global.Context, s *structures.UserEditor) *model.UserEditor {
	if s.User == nil {
		s.User = structures.DeletedUser
	}

	return &model.UserEditor{
		ID:          s.ID,
		Permissions: int(s.Permissions),
		Visible:     s.Visible,
		AddedAt:     s.AddedAt,
		User:        UserStructureToPartialModel(ctx, UserStructureToModel(ctx, s.User)),
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
		Platform:    model.ConnectionPlatform(s.Platform),
		LinkedAt:    s.LinkedAt,
		EmoteSlots:  int(s.EmoteSlots),
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
		Invisible: s.Invisible,
		Members:   []*model.User{},
	}
}

func EmoteStructureToModel(ctx global.Context, s *structures.Emote) *model.Emote {
	images := []*model.Image{}
	versions := []*model.EmoteVersion{}
	lifecycle := structures.EmoteLifecycleDisabled
	animated := false
	for _, ver := range s.Versions {
		lifecycle = ver.State.Lifecycle
		animated = ver.FrameCount > 1
		previewURL := ""
		for _, f := range ver.Formats {
			for fi, im := range f.Files {
				format := model.ImageFormatWebp
				switch f.Name {
				case structures.EmoteFormatNameAVIF:
					format = model.ImageFormatAvif
				case structures.EmoteFormatNameGIF:
					format = model.ImageFormatGif
				case structures.EmoteFormatNamePNG:
					format = model.ImageFormatPng
				}

				// Set 3x as preview
				url := fmt.Sprintf("//%s/emote/%s/%s", ctx.Config().CdnURL, ver.ID.Hex(), im.Name)
				if fi == 2 && f.Name == structures.EmoteFormatNameWEBP {
					previewURL = url
				}

				if ver.ID == s.ID {
					images = append(images, &model.Image{
						Name:     im.Name,
						Format:   format,
						URL:      url,
						Width:    int(im.Width),
						Height:   int(im.Height),
						Animated: im.Animated,
						Time:     int(im.ProcessingTime),
						Length:   int(im.Length),
					})
				}
			}
		}
		versions = append(versions, EmoteVersionStructureToModel(ctx, ver, previewURL))
	}

	owner := structures.DeletedUser
	if s.Owner != nil {
		owner = s.Owner
	}
	return &model.Emote{
		ID:        s.ID,
		Name:      s.Name,
		Flags:     int(s.Flags),
		Lifecycle: int(lifecycle),
		Tags:      s.Tags,
		Animated:  animated,
		CreatedAt: s.ID.Timestamp(),
		OwnerID:   s.OwnerID,
		Owner:     UserStructureToModel(ctx, owner),
		Channels:  &model.UserSearchResult{},
		Images:    images,
		Versions:  versions,
		Reports:   []*model.Report{},
	}
}

func EmoteStructureToPartialModel(ctx global.Context, m *model.Emote) *model.EmotePartial {
	return &model.EmotePartial{
		ID:        m.ID,
		Name:      m.Name,
		Flags:     m.Flags,
		Lifecycle: m.Lifecycle,
		Tags:      m.Tags,
		Animated:  m.Animated,
		CreatedAt: m.CreatedAt,
		OwnerID:   m.OwnerID,
		Owner:     m.Owner,
		Images:    m.Images,
	}
}

func EmoteSetStructureToModel(ctx global.Context, s *structures.EmoteSet) *model.EmoteSet {
	emotes := make([]*model.ActiveEmote, len(s.Emotes))
	for i, e := range s.Emotes {
		if e.Emote == nil {
			e.Emote = structures.DeletedEmote
		}
		emotes[i] = &model.ActiveEmote{
			ID:        e.ID,
			Name:      e.Name,
			Flags:     int(e.Flags),
			Timestamp: e.Timestamp,
			Emote:     EmoteStructureToModel(ctx, e.Emote),
		}
	}
	var owner *model.User
	if s.Owner != nil {
		owner = UserStructureToModel(ctx, s.Owner)
	}

	return &model.EmoteSet{
		ID:         s.ID,
		Name:       s.Name,
		Tags:       s.Tags,
		Emotes:     emotes,
		EmoteSlots: int(s.EmoteSlots),
		OwnerID:    &s.OwnerID,
		Owner:      owner,
	}
}

func EmoteVersionStructureToModel(ctx global.Context, s *structures.EmoteVersion, previewURL string) *model.EmoteVersion {
	return &model.EmoteVersion{
		ID:           s.ID,
		Name:         s.Name,
		Description:  s.Description,
		Timestamp:    s.ID.Timestamp(),
		ThumbnailURL: previewURL,
	}
}

func ActiveEmoteStructureToModel(ctx global.Context, s *structures.ActiveEmote) *model.ActiveEmote {
	return &model.ActiveEmote{
		ID:        s.ID,
		Name:      s.Name,
		Flags:     int(s.Flags),
		Timestamp: s.Timestamp,
	}
}
