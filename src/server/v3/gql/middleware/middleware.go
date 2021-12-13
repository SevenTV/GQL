package middleware

import (
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/src/global"
)

func New(ctx global.Context) generated.DirectiveRoot {
	return generated.DirectiveRoot{
		HasPermissions: hasPermission(ctx),
	}
}
