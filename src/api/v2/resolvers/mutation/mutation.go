package mutation

import (
	"github.com/SevenTV/GQL/graph/v2/generated"
	"github.com/SevenTV/GQL/src/api/v2/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.MutationResolver {
	return &Resolver{
		Resolver: r,
	}
}
