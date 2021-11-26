package mutation

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) CreateBan(ctx context.Context, args struct {
	VictimID string
	Reason   string
	Effects  []structures.BanEffect
	ExpireAt *string
}) (*query.BanResolver, error) {
	// Get the actor user
	actor := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrUnauthorized
	}
	if !actor.HasPermission(structures.RolePermissionManageBans) {
		return nil, helpers.ErrAccessDenied
	}

	// Find the victim
	victim, err := FetchUserWithRoles(r.Ctx, ctx, args.VictimID)
	if err != nil {
		return nil, err
	}

	// Check permissions:
	// can the actor ban this user?
	if victim.GetHighestRole().Position >= actor.GetHighestRole().Position {
		return nil, helpers.ErrAccessDenied
	}

	// Parse expiry date
	expireAt := time.Time{}
	if args.ExpireAt != nil {
		expireAt, err = time.Parse(time.RFC3339, *args.ExpireAt)
		if err != nil {
			return nil, err
		}
	}

	// Create the ban
	bb := structures.NewBanBuilder(&structures.Ban{})
	bb.Ban.ID = primitive.NewObjectID()
	bb.SetVictimID(victim.ID).
		SetActorID(actor.ID).
		SetReason(args.Reason).
		SetExpireAt(expireAt).
		SetEffects(args.Effects)

	// Write to DB
	if _, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameBans).InsertOne(ctx, bb.Ban); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	return query.CreateBanResolver(r.Ctx, ctx, bb.Ban, &bb.Ban.ID, query.GenerateSelectedFieldMap(ctx).Children)
}
