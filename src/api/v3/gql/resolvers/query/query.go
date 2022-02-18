package query

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/auth"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.QueryResolver {
	return &Resolver{r}
}

func (r *Resolver) CurrentUser(ctx context.Context) (*model.User, error) {
	actor := auth.For(ctx)
	if actor == nil {
		return nil, nil
	}

	return loaders.For(ctx).UserByID.Load(actor.ID)
}

func (r *Resolver) User(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(id)
}

func (r *Resolver) Users(ctx context.Context, query string) ([]*model.User, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) Roles(ctx context.Context) ([]*model.Role, error) {
	roles, _ := r.Ctx.Inst().Query.Roles(ctx, bson.M{})

	result := make([]*model.Role, len(roles))
	for i, rol := range roles {
		result[i] = helpers.RoleStructureToModel(r.Ctx, rol)
	}
	return result, nil
}

func (r *Resolver) Role(ctx context.Context, id primitive.ObjectID) (*model.Role, error) {
	// TODO
	return loaders.For(ctx).RoleByID.Load(id)
}

func (r *Resolver) Reports(ctx context.Context, status *model.ReportStatus, limit *int, afterID *string, beforeID *string) ([]*model.Report, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) Report(ctx context.Context, id primitive.ObjectID) (*model.Report, error) {
	return loaders.For(ctx).ReportByID.Load(id)
}

func (r *Resolver) Inbox(ctx context.Context, afterID *primitive.ObjectID) ([]*model.Message, error) {
	// TODO
	return nil, nil
}

type Sort struct {
	Value string    `json:"value"`
	Order SortOrder `json:"order"`
}

type SortOrder string

var (
	SortOrderAscending  SortOrder = "ASCENDING"
	SortOrderDescending SortOrder = "DESCENDING"
)

var sortOrderMap = map[string]int32{
	string(SortOrderDescending): 1,
	string(SortOrderAscending):  -1,
}
