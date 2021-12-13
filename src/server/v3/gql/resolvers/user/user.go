package user

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/server/v3/gql/types"
	"github.com/hashicorp/go-multierror"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.UserResolver {
	return &Resolver{r}
}

func (r *Resolver) Roles(ctx context.Context, obj *model.User) ([]*model.Role, error) {
	ids := make([]string, len(obj.Roles))
	for i, v := range obj.Roles {
		ids[i] = v.ID
	}

	roles, errs := loaders.For(ctx).RoleByID.LoadAll(ids)

	return roles, multierror.Append(nil, errs...).ErrorOrNil()
}

func (r *Resolver) OwnedEmotes(ctx context.Context, obj *model.User) ([]*model.Emote, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) Connections(ctx context.Context, obj *model.User) ([]*model.UserConnection, error) {
	ids := make([]string, len(obj.Connections))
	for i, v := range obj.Roles {
		ids[i] = v.ID
	}

	connections, errs := loaders.For(ctx).ConnectionByID.LoadAll(ids)

	return connections, multierror.Append(nil, errs...).ErrorOrNil()
}

func (r *Resolver) InboxUnreadCount(ctx context.Context, obj *model.User) (int, error) {
	// TODO
	return 0, nil
}

func (r *Resolver) Reports(ctx context.Context, obj *model.User) ([]*model.Report, error) {
	return loaders.For(ctx).ReportsByUserID.Load(obj.ID)
}
