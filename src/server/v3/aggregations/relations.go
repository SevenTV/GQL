package aggregations

import (
	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// User Relations
//
// Input: User
// Adds Field: "roles" as []Role
// Output: User
var UserRelationRoles = []bson.D{
	// Step 1: Lookup ROLE entitlements matching the input user
	{{
		Key: "$lookup",
		Value: mongo.LookupWithPipeline{
			From: mongo.CollectionNameEntitlements,
			Let:  bson.M{"user_id": "$_id"},
			Pipeline: &mongo.Pipeline{
				bson.D{{
					Key: "$match",
					Value: bson.M{
						"disabled": bson.M{"$not": bson.M{"$eq": true}},
						"kind":     "ROLE",
						"$expr": bson.M{
							"$eq": bson.A{"$user_id", "$$user_id"},
						},
					},
				}},
			},
			As: "role_entitlements",
		},
	}},
	// Step 2: Update the "role_ids" field combining the original value + entitled roles
	{{
		Key: "$set",
		Value: bson.M{
			"role_ids": bson.M{
				"$concatArrays": bson.A{"$role_ids", "$role_entitlements.data.ref"},
			},
		},
	}},
	// Step 3: Unset the temporary "role_entitlements" field
	{{Key: "$unset", Value: bson.A{"role_entitlements"}}},
	// Step 4: Lookup roles matching the newly defined role IDs and output them as "roles", an array of Role
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameRoles,
			LocalField:   "role_ids",
			ForeignField: "_id",
			As:           "roles",
		},
	}},
}

// User Relations
//
// Input: User
// Adds Field: "editors" as []UserEditor with the "user" field added to each UserEditor object
// Output: User
var UserRelationEditors = []bson.D{
	// Step 1: Lookup user editors
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameUsers,
			LocalField:   "editors.id",
			ForeignField: "_id",
			As:           "editor_user",
		},
	}},
	// Step 2: iterate over editors with user index
	{{
		Key: "$unwind",
		Value: bson.M{
			"path":              "$editor_user",
			"includeArrayIndex": "user",
		},
	}},
	// Step 3: Set "user" property to each editor object in the original editors array
	{{
		Key: "$addFields",
		Value: bson.M{
			"editors.user": "$editor_user",
		},
	}},
}

// Emote Relations
//
// Input: Emote
// Adds Field: "owner" as User
// Output: Emote
func GetEmoteRelationshipOwner(opt UserRelationshipOptions) []bson.D {
	up := mongo.Pipeline{
		bson.D{{
			Key: "$match",
			Value: bson.M{
				"$expr": bson.M{"$eq": bson.A{"$_id", "$$owner_id"}},
			},
		}},
	}
	if opt.Editors {
		up = append(up, UserRelationEditors...)
	}
	if opt.Roles {
		up = append(up, UserRelationRoles...)
	}

	p := mongo.Pipeline{
		// Step 1: Lookup emote owners
		{{
			Key: "$lookup",
			Value: mongo.LookupWithPipeline{
				From:     mongo.CollectionNameUsers,
				Let:      bson.M{"owner_id": "$owner"},
				Pipeline: &up,
				As:       "owner_user",
			},
		}},
		{{
			Key: "$set",
			Value: bson.M{
				"owner_user": bson.M{
					"$first": "$owner_user",
				},
			},
		}},
	}

	return p
}

// Emote Relations
//
// Input: User
// Adds Field: "channel_emotes" as []UserEmote with the "emote" field added to each UserEmote object
// Output: Emote
var UserRelationChannelEmotes = []bson.D{
	// Step 1: Lookup user editors
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameEmotes,
			LocalField:   "channel_emotes.id",
			ForeignField: "_id",
			As:           "_ce",
		},
	}},
	// Step 3: Set "emote" property to each UserEmote object in the original emotes array
	{{
		Key: "$set",
		Value: bson.M{
			"channel_emotes": bson.M{
				"$map": bson.M{
					"input": "$channel_emotes",
					"in": bson.M{
						"$mergeObjects": bson.A{
							"$$this",
							bson.M{
								"emote": bson.M{
									"$arrayElemAt": bson.A{
										"$_ce",
										bson.M{"$indexOfArray": bson.A{"$_ce._id", "$$this.id"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
}

type UserRelationshipOptions struct {
	Editors       bool
	Roles         bool
	ChannelEmotes bool
}
