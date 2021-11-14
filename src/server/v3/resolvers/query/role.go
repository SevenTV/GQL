package query

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleResolver struct {
	ctx context.Context
	*structures.RoleBuilder

	fields map[string]*SelectedField
	gCtx   global.Context
}

// CreateRoleResolver: generate a GQL resolver for a Role
func CreateRoleResolver(gCtx global.Context, ctx context.Context, role *structures.Role, roleID *primitive.ObjectID, fields map[string]*SelectedField) (*RoleResolver, error) {
	rb := structures.NewRoleBuilder(&structures.Role{})
	rb.Role = role

	if rb.Role == nil && roleID == nil {
		return nil, fmt.Errorf("unresolvable")
	}
	if rb.Role == nil {
		if err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameRoles).FindOne(ctx, bson.M{
			"_id": roleID,
		}).Decode(rb.Role); err != nil {
			return nil, err
		}
	}

	return &RoleResolver{
		ctx:         ctx,
		RoleBuilder: rb,
		fields:      fields,
		gCtx:        gCtx,
	}, nil
}

func (r *Resolver) Role(ctx context.Context, args struct {
	ID string
}) (*RoleResolver, error) {
	role := &structures.Role{}
	roleID, err := primitive.ObjectIDFromHex(args.ID)
	if err != nil {
		return nil, err
	}

	if err = r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameRoles).FindOne(ctx, bson.M{"_id": roleID}).Decode(role); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, helpers.ErrUnknownRole
		}

		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	return CreateRoleResolver(r.Ctx, ctx, role, &role.ID, GenerateSelectedFieldMap(ctx).Children)
}

// ID: the role's ID
func (r *RoleResolver) ID() string {
	return r.Role.ID.Hex()
}

func (r *RoleResolver) Name() string {
	return r.Role.Name
}

// Position: the role's privilege position
func (r *RoleResolver) Position() int32 {
	return r.Role.Position
}

// Color: the role's displayed color
func (r *RoleResolver) Color() int32 {
	return r.Role.Color
}

// Allowed: the role's allowed permission bits
func (r *RoleResolver) Allowed() string {
	return strconv.Itoa(int(r.Role.Allowed))
}

// Denied: the role's denied permission bits
func (r *RoleResolver) Denied() string {
	return strconv.Itoa((int(r.Role.Denied)))
}
