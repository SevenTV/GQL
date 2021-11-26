package mutation

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/structures/mutations"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) SetUserRole(ctx context.Context, args struct {
	UserID string
	RoleID string
	Action mutations.ListItemAction
}) (*query.UserResolver, error) {
	// Get the actor user
	actor, _ := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrUnauthorized
	}

	// Get the target user
	victim, err := FetchUserWithRoles(r.Ctx, ctx, args.UserID)
	if err != nil {
		return nil, err
	}

	// Get the role
	roleID, err := primitive.ObjectIDFromHex(args.RoleID)
	role := &structures.Role{}
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

	// Perform the mutation
	ub := structures.NewUserBuilder(victim)
	um := mutations.UserMutation{
		UserBuilder: ub,
	}
	if _, err = um.SetRole(ctx, r.Ctx.Inst().Mongo, mutations.SetUserRoleOptions{
		Role:   role,
		Actor:  actor,
		Action: args.Action,
	}); err != nil {
		return nil, err
	}
	// Re-fetch the victim
	victim, err = FetchUserWithRoles(r.Ctx, ctx, args.UserID)
	if err != nil {
		return nil, err
	}

	return query.CreateUserResolver(r.Ctx, ctx, victim, &victim.ID, query.GenerateSelectedFieldMap(ctx).Children)
}
