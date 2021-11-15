package query

import (
	"context"
	"sort"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *Resolver) Roles(ctx context.Context) ([]*RoleResolver, error) {
	// Check permissions
	actor := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrUnauthorized
	}
	if !actor.HasPermission(structures.RolePermissionManageRoles) {
		return nil, helpers.ErrAccessDenied
	}

	// Retrieve all roles
	roles := []*structures.Role{}
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameRoles).Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []*RoleResolver{}, nil
		}

		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	if err = cur.All(ctx, &roles); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	sort.Slice(roles, func(i, j int) bool {
		a := roles[i]
		b := roles[j]

		return a.Position > b.Position
	})

	fields := GenerateSelectedFieldMap(ctx)
	resolvers := make([]*RoleResolver, len(roles))
	for i, role := range roles {
		resolver, err := CreateRoleResolver(r.Ctx, ctx, role, &role.ID, fields.Children)
		if err != nil {
			return nil, err
		}
		resolvers[i] = resolver
	}
	return resolvers, nil
}
