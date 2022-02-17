package subscription

import (
	"context"
	"fmt"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
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
	ch := make(chan string, 1)
	ctx, cancel := context.WithCancel(ctx)

	chKey := r.Ctx.Inst().Redis.ComposeKey("events", fmt.Sprintf("sub:%s:%s", objectType, id.Hex()))
	subCh := make(chan string, 1)
	r.Ctx.Inst().Redis.Subscribe(ctx, subCh, chKey)

	go func() {
		<-ctx.Done()

		close(ch)
		close(subCh)
	}()
	go func() {
		defer cancel()

		for range subCh {
			ch <- ""
		}
	}()

	return ch
}
