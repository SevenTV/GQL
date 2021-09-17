package resolvers

import (
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
)

func Resolver() *rootResolver {
	return &rootResolver{
		Query: &query.Resolver{},
	}
}

type rootResolver struct {
	Query *query.Resolver
}
