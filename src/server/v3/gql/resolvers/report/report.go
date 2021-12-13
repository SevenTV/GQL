package report

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

func New(r types.Resolver) generated.ReportResolver {
	return &Resolver{r}
}

func (r *Resolver) Reporter(ctx context.Context, obj *model.Report) (*model.User, error) {
	return loaders.For(ctx).UserByID.Load(obj.Reporter.ID)
}

func (r *Resolver) Assignees(ctx context.Context, obj *model.Report) ([]*model.User, error) {
	ids := make([]string, len(obj.Assignees))
	for i, v := range obj.Assignees {
		ids[i] = v.ID
	}

	users, errs := loaders.For(ctx).UserByID.LoadAll(ids)

	return users, multierror.Append(nil, errs...).ErrorOrNil()
}