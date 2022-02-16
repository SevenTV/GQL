package subscription

import (
	"context"
	"fmt"

	"github.com/SevenTV/GQL/src/global"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ChangeStream(ctx global.Context) error {
	str, err := ctx.Inst().Mongo.RawDatabase().Watch(
		ctx,
		mongo.Pipeline{{{
			Key:   "$project",
			Value: bson.M{"documentKey._id": 1, "ns": 1},
		}}},
		options.ChangeStream().SetFullDocument(options.Default),
	)
	if err != nil {
		return err
	}

	go func() {
		var cd bson.M
		for str.Next(ctx) {
			if err := str.Decode(&cd); err != nil {
				logrus.WithError(err).Error("failed to decode a changestream document")
			}

			ns, ok1 := cd["ns"].(bson.M)
			k, ok2 := cd["documentKey"].(bson.M)
			if ok1 && ok2 {
				id := k["_id"].(primitive.ObjectID)
				col := ns["coll"]

				rKey := ctx.Inst().Redis.ComposeKey("gql-v3", fmt.Sprintf("sub:%s:%s", col, id.Hex()))
				ctx.Inst().Redis.RawClient().Publish(ctx, rKey.String(), "")
			}
		}
	}()

	return nil
}

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
