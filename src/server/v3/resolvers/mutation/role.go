package mutation

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/structures/mutations"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) CreateRole(ctx context.Context, args struct {
	Data CreateRoleInput
}) (*query.RoleResolver, error) {
	// Get the actor user
	actor, _ := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrUnauthorized
	}

	// Set up a new RoleBuilder & assign input data
	rb := structures.NewRoleBuilder(&structures.Role{})
	rb.Role.Name = args.Data.Name
	rb.Role.Color = args.Data.Color
	a, err := strconv.Atoi(args.Data.Allowed)
	if err != nil {
		return nil, helpers.ErrBadInt
	}
	d, err := strconv.Atoi(args.Data.Denied)
	if err != nil {
		return nil, helpers.ErrBadInt
	}
	rb.Role.Allowed = structures.RolePermission(a)
	rb.Role.Denied = structures.RolePermission(d)

	// Create the role
	rm := mutations.RoleMutation{
		RoleBuilder: rb,
	}

	_, err = rm.Create(ctx, r.Ctx.Inst().Mongo, mutations.RoleMutationOptions{
		Actor: actor,
	})
	if err != nil {
		return nil, err
	}

	return query.CreateRoleResolver(r.Ctx, ctx, rm.RoleBuilder.Role, &rm.RoleBuilder.Role.ID, query.GenerateSelectedFieldMap(ctx).Children)
}

func (r *Resolver) EditRole(ctx context.Context, args struct {
	RoleID string
	Data   EditRoleInput
}) (*query.RoleResolver, error) {
	// Get the actor user
	actor, _ := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrUnauthorized
	}

	// Find the role
	role := &structures.Role{}
	roleID, err := primitive.ObjectIDFromHex(args.RoleID)
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

	// Apply mutations
	ogPos := role.Position
	rb := structures.NewRoleBuilder(role)
	if args.Data.Name != nil {
		name := *args.Data.Name
		rb.SetName(name)
	}
	if args.Data.Color != nil {
		color := *args.Data.Color
		rb.SetColor(color)
	}
	if args.Data.Position != nil {
		pos := *args.Data.Position
		if pos < 0 {
			return nil, fmt.Errorf("bad role position (expected 0 or higher but got %d)", pos)
		}
		rb.SetPosition(pos)
	}
	if args.Data.Allowed != nil {
		allowed := *args.Data.Allowed
		a, err := strconv.Atoi(allowed)
		if err != nil {
			return nil, helpers.ErrBadInt
		}
		rb.SetAllowed(structures.RolePermission(a))
	}
	if args.Data.Denied != nil {
		denied := *args.Data.Denied
		d, err := strconv.Atoi(denied)
		if err != nil {
			return nil, helpers.ErrBadInt
		}
		rb.SetAllowed(structures.RolePermission(d))
	}

	// Update the role
	rm := mutations.RoleMutation{
		RoleBuilder: rb,
	}
	if _, err := rm.Edit(ctx, r.Ctx.Inst().Mongo, mutations.RoleEditOptions{
		Actor:            actor,
		OriginalPosition: ogPos,
	}); err != nil {
		return nil, err
	}

	return query.CreateRoleResolver(r.Ctx, ctx, role, &role.ID, query.GenerateSelectedFieldMap(ctx).Children)
}

func (r *Resolver) DeleteRole(ctx context.Context, args struct {
	RoleID string
}) (string, error) {
	// Get the actor user
	actor, _ := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return "", helpers.ErrUnauthorized
	}

	// Find the role
	role := &structures.Role{}
	roleID, err := primitive.ObjectIDFromHex(args.RoleID)
	if err != nil {
		return "", err
	}
	if err = r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameRoles).FindOne(ctx, bson.M{"_id": roleID}).Decode(role); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", helpers.ErrUnknownRole
		}

		logrus.WithError(err).Error("mongo")
		return "", err
	}

	// Delete the role
	rm := mutations.RoleMutation{
		RoleBuilder: structures.NewRoleBuilder(role),
	}
	if _, err = rm.Delete(ctx, r.Ctx.Inst().Mongo, mutations.RoleMutationOptions{
		Actor: actor,
	}); err != nil {
		return "", err
	}

	return role.ID.Hex(), nil
}

type CreateRoleInput struct {
	Name    string `json:"name"`
	Color   int32  `json:"color"`
	Allowed string `json:"allowed"`
	Denied  string `json:"denied"`
}

type EditRoleInput struct {
	Name     *string `json:"name"`
	Color    *int32  `json:"color"`
	Position *int32  `json:"position"`
	Allowed  *string `json:"allowed"`
	Denied   *string `json:"denied"`
}
