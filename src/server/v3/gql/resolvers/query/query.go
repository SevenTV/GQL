package query

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/server/v3/gql/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.QueryResolver {
	return &Resolver{r}
}

func (r *Resolver) User(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(id)
}

func (r *Resolver) Users(ctx context.Context, query string) ([]*model.User, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) Emote(ctx context.Context, id primitive.ObjectID) (*model.Emote, error) {
	return loaders.For(ctx).EmoteByID.Load(id)
}

func (r *Resolver) Emotes(ctx context.Context, query string, page *int, limit *int, filter *model.EmoteFilter, sort *model.EmoteSortOrder) ([]*model.Emote, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) Roles(ctx context.Context) ([]*model.Role, error) {
	// TODO
	return nil, nil
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
