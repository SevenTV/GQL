package events

import (
	"fmt"

	"github.com/SevenTV/GQL/src/global"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Publish(ctx global.Context, objectType string, id primitive.ObjectID) {
	k := ctx.Inst().Redis.ComposeKey("events", fmt.Sprintf("sub:%s:%s", objectType, id.Hex()))
	ctx.Inst().Redis.RawClient().Publish(ctx, k.String(), "1")
}
