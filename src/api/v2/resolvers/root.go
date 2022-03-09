package resolvers

import (
	"github.com/SevenTV/GQL/graph/v2/generated"
	"github.com/SevenTV/GQL/src/api/v2/resolvers/query"
	"github.com/SevenTV/GQL/src/api/v2/resolvers/user"
	"github.com/SevenTV/GQL/src/api/v2/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.ResolverRoot {
	return &Resolver{
		Resolver: r,
	}
}

func (r *Resolver) Query() generated.QueryResolver {
	return query.New(r.Resolver)
}

func (r *Resolver) User() generated.UserResolver {
	return user.New(r.Resolver)
}
