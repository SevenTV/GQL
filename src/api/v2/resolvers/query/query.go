package query

import (
	"context"

	"github.com/SevenTV/GQL/graph/v2/generated"
	"github.com/SevenTV/GQL/src/api/v2/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.QueryResolver {
	return &Resolver{r}
}

func (r *Resolver) HelloWorld(ctx context.Context) (string, error) {
	return "Hello, World!", nil
}
