package query

import (
	"context"
	"fmt"

	"github.com/SevenTV/GQL/src/structures"
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
