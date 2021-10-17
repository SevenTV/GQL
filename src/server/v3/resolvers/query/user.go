package query

import (
	"context"
	"fmt"

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
	ub := structures.NewUserBuilder()

	if user == nil && userID == nil {
		return nil, fmt.Errorf("unresolvable")
	}
	if user == nil {
		ub.User = &structures.User{}

		// Begin aggregation pipeline
		pipeline := mongo.Pipeline{
			bson.D{{
				Key: "$match",
				Value: bson.M{
					"_id": userID,
				},
			}},
		}

		// Relation: Roles
		if _, ok := fields["roles"]; ok {
			pipeline = append(pipeline, aggregations.UserRelationRoles...)
		}

		// Relation: Editors
		if _, ok := fields["editors"]; ok {

		}

		cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}

		cur.Next(ctx)
		cur.Close(ctx)
		if err = cur.Decode(ub.User); err != nil {
			return nil, err
		}
	}

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
