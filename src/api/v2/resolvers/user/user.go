package user

import (
	"context"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/v2/generated"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/helpers"
	"github.com/SevenTV/GQL/src/api/v2/loaders"
	"github.com/SevenTV/GQL/src/api/v2/types"
	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.UserResolver {
	return &Resolver{
		Resolver: r,
	}
}

func (r *Resolver) Role(ctx context.Context, obj *model.User) (*model.Role, error) {
	if obj.Role == nil {
		// Get default role
		roles, err := r.Ctx.Inst().Query.Roles(ctx, bson.M{"default": true})
		if err == nil && len(roles) > 0 {
			obj.Role = helpers.RoleStructureToModel(r.Ctx, roles[0])
		} else {
			obj.Role = helpers.RoleStructureToModel(r.Ctx, structures.NilRole)
		}
	}
	return obj.Role, nil
}

func (r *Resolver) Emotes(ctx context.Context, obj *model.User) ([]*model.Emote, error) {
	return loaders.For(ctx).UserEmotes.Load(obj.EmoteSetID)
}

func (r *Resolver) EmoteIds(ctx context.Context, obj *model.User) ([]string, error) {
	result := []string{}
	emotes, err := loaders.For(ctx).UserEmotes.Load(obj.EmoteSetID)
	if err != nil {
		return result, err
	}

	for _, e := range emotes {
		result = append(result, e.ID)
	}
	return result, nil
}

func (r *Resolver) EmoteAliases(ctx context.Context, obj *model.User) ([][]string, error) {
	result := [][]string{}
	if obj.EmoteSetID == "" {
		return result, nil
	}
	emotes, err := loaders.For(ctx).UserEmotes.Load(obj.EmoteSetID)
	if err != nil {
		return result, err
	}
	for _, e := range emotes {
		if e.OriginalName == nil {
			continue // no original name property means no alias set
		}
		result = append(result, []string{e.ID, e.Name})
	}

	return result, nil
}

func (r *Resolver) Editors(ctx context.Context, obj *model.User) ([]*model.UserPartial, error) {
	result := []*model.UserPartial{}
	editors, errs := loaders.For(ctx).UserByID.LoadAll(obj.EditorIds)
	if err := multierror.Append(nil, errs...).ErrorOrNil(); err != nil {
		return result, err
	}

	for _, ed := range editors {
		result = append(result, helpers.UserStructureToPartialModel(r.Ctx, ed))
	}
	return result, nil
}

func (r *Resolver) EditorIn(ctx context.Context, obj *model.User) ([]*model.UserPartial, error) {
	result := []*model.UserPartial{}
	userID, err := primitive.ObjectIDFromHex(obj.ID)
	if err != nil {
		return result, err
	}

	editors, err := r.Ctx.Inst().Query.UserEditorOf(ctx, userID)
	if err != nil {
		return result, err
	}

	// Get a list of user IDs from the v3 editor list
	ids := make([]string, len(editors))
	for i, ed := range editors {
		ids[i] = ed.ID.Hex()
	}

	users, errs := loaders.For(ctx).UserByID.LoadAll(ids)
	if err = multierror.Append(nil, errs...).ErrorOrNil(); err != nil {
		return result, err
	}
	for _, u := range users {
		result = append(result, helpers.UserStructureToPartialModel(r.Ctx, u))
	}
	return result, nil
}
