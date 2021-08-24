package resolvers

import (
	"github.com/SevenTV/ThreeLetterAPI/src/server/v3/gql/resolvers/query"
)

func Resolver() *rootResolver {
	return &rootResolver{
		Query: &query.Resolver{},
	}
}

type rootResolver struct {
	Query *query.Resolver
}
