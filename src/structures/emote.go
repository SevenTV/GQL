package structures

import (
	"context"

	"github.com/SevenTV/GQL/src/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmoteBuilder: Wraps an Emote and offers methods to fetch and mutate emote data
type EmoteBuilder struct {
	Emote *Emote
}

// FetchByID: Get an emote by its ID
func (b EmoteBuilder) FetchByID(ctx context.Context, id primitive.ObjectID) (*Emote, error) {
	doc := mongo.Collection(mongo.CollectionNameEmotes).FindOne(ctx, bson.M{
		"_id": id,
	})
	if err := doc.Err(); err != nil {
		return nil, err
	}

	var emote Emote
	if err := doc.Decode(&emote); err != nil {
		return nil, err
	}

	return &emote, nil
}

type Emote struct {
	ID   primitive.ObjectID `json:"id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}

const (
	EmoteStatusDeleted int32 = iota - 1
	EmoteStatusProcessing
	EmoteStatusPending
	EmoteStatusDisabled
	EmoteStatusLive
)

const (
	EmoteVisibilityPrivate int32 = 1 << iota
	EmoteVisibilityGlobal
	EmoteVisibilityUnlisted
	EmoteVisibilityOverrideBTTV
	EmoteVisibilityOverrideFFZ
	EmoteVisibilityOverrideTwitchGlobal
	EmoteVisibilityOverrideTwitchSubscriber
	EmoteVisibilityZeroWidth

	EmoteVisibilityAll int32 = (1 << iota) - 1
)
