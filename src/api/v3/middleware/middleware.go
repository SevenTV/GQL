package middleware

import (
	"github.com/SevenTV/GQL/graph/v3/generated"
	"github.com/SevenTV/GQL/src/global"
)

func New(ctx global.Context) generated.DirectiveRoot {
	return generated.DirectiveRoot{
		HasPermissions: hasPermission(ctx),
		Internal:       internal(ctx),
	}
}
