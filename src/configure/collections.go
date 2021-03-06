package configure

import (
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Indexes = []mongo.IndexRef{
	{
		Collection: mongo.CollectionNameEmoteSets,
		Index: mongo.IndexModel{
			Keys: bson.M{"num_id": 1},
			Options: &options.IndexOptions{
				Unique: utils.BoolPointer(true),
			},
		},
	},
	{
		Collection: mongo.CollectionNameEmoteSets,
		Index: mongo.IndexModel{
			Keys: bson.M{"emote_ids": 1},
		},
	},
}
