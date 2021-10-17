package query

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserResolver struct {
	ctx context.Context
	*structures.UserBuilder

	fields map[string]*SelectedField
	gCtx   global.Context
}

func CreateUserResolver(gCtx global.Context, ctx context.Context, user *structures.User, userID *primitive.ObjectID, fields map[string]*SelectedField) (*UserResolver, error) {
	ub := structures.NewUserBuilder()
	ub.User = user

	if ub.User == nil && userID == nil {
		return nil, fmt.Errorf("unresolvable")
	}
	if ub.User == nil {
		doc := gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
			"_id": userID,
		})
		if err := doc.Decode(&user); err != nil {
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

func (r *UserResolver) ID() string {
	return r.User.ID.Hex()
}

func (r *UserResolver) Username() string {
	return r.User.Username
}

func (r *UserResolver) DisplayName() string {
	return r.User.Username
}
