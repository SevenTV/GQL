package loaders

import (
	"context"
	"time"

	"github.com/SevenTV/GQL/graph/v2/loaders"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/helpers"
	"github.com/SevenTV/GQL/src/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func userEmotesLoader(gCtx global.Context) *loaders.UserEmotesLoader {
	return loaders.NewUserEmotesLoader(loaders.UserEmotesLoaderConfig{
		Wait: time.Millisecond * 25,
		Fetch: func(keys []string) ([][]*model.Emote, []error) {
			ctx, cancel := context.WithTimeout(gCtx, time.Second*10)
			defer cancel()

			modelLists := make([][]*model.Emote, len(keys))
			errs := make([]error, len(keys))

			ids := make([]primitive.ObjectID, len(keys))
			for i, k := range keys {
				ids[i], _ = primitive.ObjectIDFromHex(k)
			}

			sets, err := gCtx.Inst().Query.EmoteSets(ctx, bson.M{"_id": bson.M{"$in": ids}})
			if err == nil {
				m := make(map[primitive.ObjectID][]*model.Emote)
				// iterate over sets
				for _, set := range sets {
					// iterate over emotes of set
					for _, ae := range set.Emotes {
						em := helpers.EmoteStructureToModel(gCtx, ae.Emote)

						// set "alias"?
						if ae.Name != em.Name {
							em.OriginalName = &ae.Emote.Name
							em.Name = ae.Name
						}

						m[set.ID] = append(m[set.ID], em)
					}
				}

				for i, v := range ids {
					if x, ok := m[v]; ok {
						modelLists[i] = x
					}
				}
			}

			return modelLists, errs
		},
	})
}
