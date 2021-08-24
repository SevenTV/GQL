package query

import (
	"context"

	"github.com/SevenTV/ThreeLetterAPI/src/structures"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func EmoteResolver(ctx context.Context, emote *structures.Emote, emoteID *primitive.ObjectID, fields map[string]*SelectedField) error {
	return nil
}
