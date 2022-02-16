package user

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/mutations"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/auth"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
	"go.mongodb.org/mongo-driver/bson"
)

type ResolverOps struct {
	types.Resolver
}

func NewOps(r types.Resolver) generated.UserOpsResolver {
	return &ResolverOps{r}
}

func (r *ResolverOps) Connections(ctx context.Context, obj *model.UserOps, id string, d model.UserConnectionUpdate) ([]*model.UserConnection, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, errors.ErrUnauthorized()
	}

	b := structures.NewUserBuilder(nil)
	if err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
		"_id": obj.ID,
	}).Decode(b.User); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrUnknownUser()
		}
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Perform a mutation
	var err error
	m := mutations.UserMutation{UserBuilder: b}
	if d.EmoteSetID != nil {
		if _, err = m.SetActiveEmoteSet(ctx, r.Ctx.Inst().Mongo, mutations.SetUserActiveEmoteSet{
			EmoteSetID:   *d.EmoteSetID,
			Platform:     structures.UserConnectionPlatformTwitch,
			Actor:        actor,
			ConnectionID: id,
		}); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	result := helpers.UserStructureToModel(r.Ctx, b.User)
	return result.Connections, nil
}
