package mutation

import (
	"context"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	Ctx global.Context
}

// FetchUserWithRoles: Fetches a user from their ID and adds role data for permission checks
func FetchUserWithRoles(gCtx global.Context, ctx context.Context, userID string) (*structures.User, error) {
	usr := &structures.User{}
	usrID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, helpers.ErrBadObjectID
	}
	cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, append(mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": usrID}}},
	}, aggregations.UserRelationRoles...))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, helpers.ErrUnknownUser
		}

		logrus.WithError(err).Error("mongo")
		return nil, helpers.ErrInternalServerError
	}
	cur.Next(ctx)
	if err = cur.Decode(usr); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	cur.Close(ctx)

	return usr, nil
}
