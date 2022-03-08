package subscription

import (
	"context"
	"fmt"

	"github.com/SevenTV/GQL/graph/v3/generated"
	"github.com/SevenTV/GQL/src/api/v3/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.SubscriptionResolver {
	return &Resolver{
		Resolver: r,
	}
}

func (r *Resolver) subscribe(ctx context.Context, objectType string, id primitive.ObjectID) <-chan string {
	chKey := r.Ctx.Inst().Redis.ComposeKey("events", fmt.Sprintf("sub:%s:%s", objectType, id.Hex()))
	subCh := make(chan string, 1)
	r.Ctx.Inst().Redis.Subscribe(ctx, subCh, chKey)

	go func() {
		<-ctx.Done()
		close(subCh)
	}()

	return subCh
}
