package query

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmoteResolver struct {
	ctx context.Context
	*structures.EmoteBuilder

	fields map[string]*SelectedField
}

func CreateEmoteResolver(ctx context.Context, emote *structures.Emote, emoteID *primitive.ObjectID, fields map[string]*SelectedField) (*EmoteResolver, error) {
	eb := structures.EmoteBuilder{Emote: emote}

	if eb.Emote == nil && emoteID == nil {
		return nil, fmt.Errorf("Unresolvable")
	}
	if eb.Emote == nil {
		if _, err := eb.FetchByID(ctx, *emoteID); err != nil {
			return nil, err
		}
	}

	return &EmoteResolver{
		ctx,
		&eb,
		fields,
	}, nil
}

// ID: resolves the ID of the emote
func (r *EmoteResolver) ID() string {
	return r.Emote.ID.Hex()
}

// Name: resolves the name of the emote
func (r *EmoteResolver) Name() string {
	return r.Emote.Name
}

// URLs: resolves a list of cdn urls for the emote
func (r *EmoteResolver) URLs() [][]string {
	result := make([][]string, 4) // 4 length because there are 4 CDN sizes supported (1x, 2x, 3x, 4x)

	for i := 1; i <= 4; i++ {
		a := make([]string, 2)
		a[0] = fmt.Sprintf("%d", i)
		a[1] = utils.GetCdnURL(r.Emote.ID.Hex(), int8(i))

		result[i-1] = a
	}

	r.Emote.URLs = result
	return r.Emote.URLs
}
