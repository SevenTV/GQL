package subscription

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *Resolver) subscribe(ctx context.Context, objectType string, id primitive.ObjectID) <-chan string {
	ch := make(chan string, 1)
	ctx, cancel := context.WithCancel(ctx)

	chKey := r.Ctx.Inst().Redis.ComposeKey("gql-v3", fmt.Sprintf("sub:%s:%s", objectType, id.Hex()))
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
