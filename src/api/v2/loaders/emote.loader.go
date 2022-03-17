package loaders

import (
	"context"
	"time"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/v2/loaders"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/helpers"
	"github.com/SevenTV/GQL/src/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func emoteByID(gCtx global.Context) *loaders.EmoteLoader {
	return loaders.NewEmoteLoader(loaders.EmoteLoaderConfig{
		Wait: time.Millisecond * 25,
		Fetch: func(keys []string) ([]*model.Emote, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			// Fetch emote data from the database
			models := make([]*model.Emote, len(keys))
			errs := make([]error, len(keys))

			// Parse object IDs
			ids := make([]primitive.ObjectID, len(keys))
			for i, k := range keys {
				id, err := primitive.ObjectIDFromHex(k)
				if err != nil {
					errs[i] = err
					continue
				}
				ids[i] = id
			}

			// Fetch emotes
			emotes, err := gCtx.Inst().Query.Emotes(ctx, bson.M{
				"versions.id": bson.M{"$in": ids},
			})

			if err == nil {
				m := make(map[primitive.ObjectID]*structures.Emote)
				for _, e := range emotes {
					if e == nil {
						continue
					}
					for _, ver := range e.Versions {
						m[ver.ID] = e
					}
				}

				for i, v := range ids {
					if x, ok := m[v]; ok {
						ver, _ := x.GetVersion(v)
						if ver == nil || ver.IsUnavailable() {
							continue
						}
						x.ID = v
						models[i] = helpers.EmoteStructureToModel(gCtx, x)
					}
				}
			}

			return models, errs
		},
	})
}
