package query

import (
	"context"
	"strconv"
	"strings"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/query"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/helpers"
	"github.com/SevenTV/GQL/src/api/v2/loaders"
	"github.com/SevenTV/GQL/src/api/v3/auth"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) User(ctx context.Context, id string) (*model.User, error) {
	if primitive.IsValidObjectID(id) {
		return loaders.For(ctx).UserByID.Load(id)
	} else if id == "@me" {
		// Handle @me (fetch actor)
		// this sets the queried user ID to that of the actor user
		actor := auth.For(ctx)
		id = actor.ID.Hex()
		return loaders.For(ctx).UserByID.Load(id)
	} else {
		// at this point we assume the query is for a username
		// (it was neither an id, or the @me label)
		return loaders.For(ctx).UserByUsername.Load(strings.ToLower(id))
	}
}

func (r *Resolver) SearchUsers(ctx context.Context, queryArg string, page *int, limit *int) ([]*model.UserPartial, error) {
	actor := auth.For(ctx)
	if actor == nil || !actor.HasPermission(structures.RolePermissionManageUsers) {
		return nil, errors.ErrInsufficientPrivilege()
	}
	users, totalCount, err := r.Ctx.Inst().Query.SearchUsers(ctx, bson.M{}, query.UserSearchOptions{
		Page:  1,
		Limit: 250,
		Query: queryArg,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*model.UserPartial, len(users))
	for i, u := range users {
		result[i] = helpers.UserStructureToPartialModel(r.Ctx, helpers.UserStructureToModel(r.Ctx, u))
	}

	rctx := ctx.Value(helpers.RequestCtxKey).(*fasthttp.RequestCtx)
	if rctx != nil {
		rctx.Response.Header.Set("X-Collection-Size", strconv.Itoa(totalCount))
	}
	return result, nil
}
