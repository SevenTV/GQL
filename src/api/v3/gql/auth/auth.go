package auth

import (
	"context"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
)

func For(ctx context.Context) *structures.User {
	raw, _ := ctx.Value(helpers.UserKey).(*structures.User)
	return raw
}
