package mutation

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *Resolver) CreateBan(ctx context.Context, args struct {
	VictimID string
	Reason   string
	Effects  []structures.BanEffect
	ExpireAt *string
}) (*query.BanResolver, error) {
	// Get the actor user
	actor, _ := ctx.Value(helpers.UserKey).(*structures.User)
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
	if victim.ID == actor.ID { // the actor cannot ban themselve
		return nil, helpers.ErrDontBeSilly
	}
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
		logrus.WithError(err).Error("mongo, failed to write ban")
		return nil, err
	}

	// Remove the user's bound roles
	if _, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).UpdateOne(ctx, bson.M{"_id": victim.ID}, bson.M{
		"$set": bson.M{"role_ids": []primitive.ObjectID{}},
	}); err != nil {
		logrus.WithError(err).Error("mongo, failed to remove banned user's bound roles")
	}

	return query.CreateBanResolver(r.Ctx, ctx, bb.Ban, &bb.Ban.ID, query.GenerateSelectedFieldMap(ctx).Children)
}

func (r *Resolver) EditBan(ctx context.Context, args struct {
	BanID    string
	Reason   *string
	Effects  *[]structures.BanEffect
	ExpireAt *string
}) (*query.BanResolver, error) {
	// Get the actor user
	actor, _ := ctx.Value(helpers.UserKey).(*structures.User)
	if actor == nil {
		return nil, helpers.ErrUnauthorized
	}
	if !actor.HasPermission(structures.RolePermissionManageBans) {
		return nil, helpers.ErrAccessDenied
	}

	// Find the ban
	ban := &structures.Ban{}
	banID, err := primitive.ObjectIDFromHex(args.BanID)
	if err != nil {
		return nil, helpers.ErrBadObjectID
	}
	if err = r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameBans).FindOne(ctx, bson.M{"_id": banID}).Decode(ban); err != nil {
		logrus.WithError(err).Error("mongo, couldn't find the ban to edit")
		return nil, err
	}

	// Apply mutations
	bb := structures.NewBanBuilder(ban)
	if args.Reason != nil { // set reason
		bb.SetReason(*args.Reason)
	}
	if args.Effects != nil { // set effects
		bb.SetEffects(*args.Effects)
	}
	if args.ExpireAt != nil { // set expire date
		t, err := time.Parse(time.RFC3339, *args.ExpireAt)
		if err != nil {
			return nil, err
		}
		bb.SetExpireAt(t)
	}

	// Write to DB
	if err = r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameBans).FindOneAndUpdate(ctx, bson.M{
		"_id": ban.ID,
	}, bb.Update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(ban); err != nil {
		logrus.WithError(err).Error("mongo, failed to write ban update")
		return nil, err
	}

	return query.CreateBanResolver(r.Ctx, ctx, ban, &ban.ID, query.GenerateSelectedFieldMap(ctx).Children)
}
