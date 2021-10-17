package mutation

import (
	"context"

	"github.com/SevenTV/GQL/src/global"
)

type Resolver struct {
	Ctx global.Context
}

func (r *Resolver) Test(ctx context.Context, args struct {
	Foobar string
}) string {
	return "foobar"
}
