package resolvers

import (
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/resolvers/query"
)

func Resolver(gCtx global.Context) *rootResolver {
	return &rootResolver{
		Query: &query.Resolver{
			Ctx: gCtx,
		},
		Ctx: gCtx,
	}
}

type rootResolver struct {
	Query *query.Resolver

	Ctx global.Context
}
