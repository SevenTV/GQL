package query

import (
	"context"
	"fmt"
	"sort"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/aggregations"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserResolver struct {
	ctx context.Context
	*structures.UserBuilder

	fields map[string]*SelectedField
	gCtx   global.Context
}

// CreateUserResolver: generate a GQL resolver for a User
func CreateUserResolver(gCtx global.Context, ctx context.Context, user *structures.User, userID *primitive.ObjectID, fields map[string]*SelectedField) (*UserResolver, error) {
	if user == nil && userID == nil {
		return nil, fmt.Errorf("unresolvable")
	}

	var pipeline mongo.Pipeline
	if user == nil {
		pipeline = mongo.Pipeline{
			{{
				Key: "$match",
				Value: bson.M{
					"_id": userID,
				},
			}},
		}
		user = &structures.User{}
	} else {
		pipeline = mongo.Pipeline{
			{{
				Key: "$replaceRoot",
				Value: bson.M{
					"newRoot": user,
				},
			}},
		}
	}

	// Relation: Roles
	if _, ok := fields["roles"]; ok && user.Roles == nil {
		pipeline = append(pipeline, aggregations.UserRelationRoles...)
	}
	if _, ok := fields["tag_color"]; ok && user.Roles == nil {
		if _, ok = fields["roles"]; !ok {
			pipeline = append(pipeline, aggregations.UserRelationRoles...)
		}
	}

	// Relation: Editors
	if _, ok := fields["editors"]; ok && user.Editors == nil {
		pipeline = append(pipeline, aggregations.UserRelationEditors...)
	}

	// Relation: Channel Emotes
	if _, ok := fields["channel_emotes"]; ok {
		fetched := true
		if user.ChannelEmotes == nil {
			fetched = false
		} else {
			for _, ce := range user.ChannelEmotes {
				if ce.Emote == nil {
					fetched = false
					break
				}
			}
		}

		if !fetched {
			pipeline = append(pipeline, aggregations.UserRelationChannelEmotes...)
		}
	}

	if len(pipeline) > 1 {
		cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		cur.Next(ctx)
		cur.Close(ctx)
		if err = cur.Decode(user); err != nil {
			return nil, err
		}
	}

	ub := structures.NewUserBuilder(user)
	return &UserResolver{
		ctx:         ctx,
		UserBuilder: ub,
		fields:      fields,
		gCtx:        gCtx,
	}, nil
}

func (r *Resolver) User(ctx context.Context, args struct {
	ID string
}) (*UserResolver, error) {
	user, ok := ctx.Value(utils.Key("user")).(*structures.User)

	var (
		resolver *UserResolver
		err      error
	)
	fields := GenerateSelectedFieldMap(ctx)
	if args.ID == "@me" && ok {
		resolver, err = CreateUserResolver(r.Ctx, ctx, user, &user.ID, fields.Children)
		if err != nil {
			return nil, err
		}
	} else {
		id, err := primitive.ObjectIDFromHex(args.ID)
		if err != nil {
			return nil, err
		}

		resolver, err = CreateUserResolver(r.Ctx, ctx, nil, &id, fields.Children)
		if err != nil {
			return nil, err
		}
	}

	return resolver, nil
}

// ID: the user's ID
func (r *UserResolver) ID() string {
	return r.User.ID.Hex()
}

// UserType: the type of user account (i.e BOT, SYSTEM)
func (r *UserResolver) UserType() string {
	return string(r.User.UserType)
}

// Username: the username
func (r *UserResolver) Username() string {
	return r.User.Username
}

// DisplayName: the user's display name
func (r *UserResolver) DisplayName() string {
	return r.User.Username
}

// AvatarURL: an HTTP URL to the user's avatar
func (r *UserResolver) AvatarURL() string {
	return r.User.AvatarURL
}

// Biography: a short description for the user
func (r *UserResolver) Biography() string {
	return r.User.Biography
}

// Role: user's role
func (r *UserResolver) Roles() ([]*RoleResolver, error) {
	resolvers := make([]*RoleResolver, len(r.User.Roles))

	fields := GenerateSelectedFieldMap(r.ctx)
	for i, role := range r.User.Roles {
		resolver, err := CreateRoleResolver(r.gCtx, r.ctx, role, &role.ID, fields.Children)
		if err != nil {
			return nil, err
		}

		resolvers[i] = resolver
	}

	return resolvers, nil
}

func (r *UserResolver) TagColor() int32 {
	if len(r.User.Roles) == 0 {
		return 0
	}

	sort.Slice(r.User.Roles, func(i, j int) bool {
		a := r.User.Roles[i]
		b := r.User.Roles[j]

		return a.Position > b.Position
	})

	return r.User.Roles[0].Color
}

func (r *UserResolver) Editors() ([]*UserEditorResolvable, error) {
	result := make([]*UserEditorResolvable, len(r.User.Editors))

	fields := GenerateSelectedFieldMap(r.ctx)
	for i, editor := range r.User.Editors {
		if editor.User == nil {
			continue
		}

		ur, err := CreateUserResolver(r.gCtx, r.ctx, editor.User, &editor.User.ID, fields.Children)
		if err != nil {
			return nil, err
		}

		result[i] = &UserEditorResolvable{
			User:        ur,
			Connections: []string{},
			Permissions: 0,
			Visible:     true,
		}
	}

	return result, nil
}

func (r *UserResolver) ChannelEmotes() ([]*UserEmoteResolvable, error) {
	result := make([]*UserEmoteResolvable, len(r.User.ChannelEmotes))

	fields := GenerateSelectedFieldMap(r.ctx)
	for i, emote := range r.User.ChannelEmotes {
		if emote == nil {
			continue
		}

		er, err := CreateEmoteResolver(r.gCtx, r.ctx, emote.Emote, &emote.ID, fields.Children)
		if err != nil {
			return nil, err
		}

		result[i] = &UserEmoteResolvable{
			Emote:       er,
			Connections: []string{},
			Alias:       "",
			ZeroWidth:   false,
		}
	}

	return result, nil
}

type UserEditorResolvable struct {
	User        *UserResolver `json:"user"`
	Connections []string      `json:"connections"`
	Permissions int32         `json:"permissions"`
	Visible     bool          `json:"visible"`
}

type UserEmoteResolvable struct {
	Emote       *EmoteResolver `json:"emote"`
	Connections []string       `json:"connections,omitempty"`
	Alias       string         `json:"alias,omitempty"`
	ZeroWidth   bool           `json:"zero_width,omitempty"`
}
