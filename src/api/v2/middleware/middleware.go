package middleware

import (
	"github.com/SevenTV/GQL/graph/v2/generated"
	"github.com/SevenTV/GQL/src/global"
)

func New(ctx global.Context) generated.DirectiveRoot {
	return generated.DirectiveRoot{
		Internal: internal(ctx),
	}
}
