package query

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmoteResolver struct {
	ctx   context.Context
	quota *helpers.Quota
	*structures.EmoteBuilder

	fields map[string]*SelectedField
	gCtx   global.Context
}

func CreateEmoteResolver(gCtx global.Context, ctx context.Context, emote *structures.Emote, emoteID *primitive.ObjectID, fields map[string]*SelectedField) (*EmoteResolver, error) {
	if emote == nil && emoteID == nil {
		return nil, fmt.Errorf("unresolvable")
	}
	var pipeline mongo.Pipeline
	if emote == nil {
		pipeline = mongo.Pipeline{
			{{Key: "$match", Value: bson.M{
				"_id":    emoteID,
				"status": bson.M{"$gte": structures.EmoteStatusProcessing},
			}}},
		}
		emote = &structures.Emote{}
	} else {
		pipeline = mongo.Pipeline{
			{{
				Key: "$replaceRoot",
				Value: bson.M{
					"newRoot": emote,
				},
			}},
		}
	}

	// Query owner sub-fields
	if of, ok := fields["owner"]; ok && (!emote.OwnerID.IsZero() && emote.Owner == nil) {
		_, qEditors := of.Children["editors"]
		_, qRoles := of.Children["roles"]
		if !qRoles {
			_, qRoles = of.Children["tag_color"]
		}
		_, qChannelEmotes := of.Children["channel_emotes"]
		opt := aggregations.UserRelationshipOptions{
			Editors:       qEditors,
			Roles:         qRoles,
			ChannelEmotes: qChannelEmotes,
		}

		pipeline = append(pipeline, aggregations.GetEmoteRelationshipOwner(opt)...)
	}

	if emote.ID.IsZero() || len(pipeline) > 0 {
		cur, _ := gCtx.Inst().Mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, pipeline)
		if ok := cur.TryNext(ctx); !ok {
			return nil, helpers.ErrUnknownEmote
		}
		if err := cur.Decode(emote); err != nil {
			logrus.WithError(err).Error("mongo")
			return nil, err
		}
		cur.Close(ctx)
	}

	eb := &structures.EmoteBuilder{Emote: emote}
	return &EmoteResolver{
		ctx:          ctx,
		quota:        ctx.Value(helpers.QuotaKey).(*helpers.Quota),
		EmoteBuilder: eb,
		fields:       fields,
		gCtx:         gCtx,
	}, nil
}

func (r *Resolver) Emote(ctx context.Context, args struct {
	ID string
}) (*EmoteResolver, error) {
	// Parse ID
	emoteID, err := primitive.ObjectIDFromHex(args.ID)
	if err != nil {
		return nil, err
	}

	fields := GenerateSelectedFieldMap(ctx)
	return CreateEmoteResolver(r.Ctx, ctx, nil, &emoteID, fields.Children)
}

// ID: resolves the ID of the emote
func (r *EmoteResolver) ID() string {
	return r.Emote.ID.Hex()
}

// Name: resolves the name of the emote
func (r *EmoteResolver) Name() string {
	return r.Emote.Name
}

// Visibility: the visibility bitfield for the emote
func (r *EmoteResolver) Flags() int32 {
	return int32(r.Emote.Flags)
}

// Status: emote status
func (r *EmoteResolver) Status() int32 {
	return int32(r.Emote.Status)
}

// Tags: emote search tags
func (r *EmoteResolver) Tags() []string {
	return r.Emote.Tags
}

// URLs: resolves a list of cdn urls for the emote
func (r *EmoteResolver) Links() [][]string {
	if ok := r.quota.DecreaseByOne("Emote", "URLs"); !ok {
		return nil
	}

	result := make([][]string, 4) // 4 length because there are 4 CDN sizes supported (1x, 2x, 3x, 4x)

	for i := 1; i <= 4; i++ {
		a := make([]string, 2)
		a[0] = fmt.Sprintf("%d", i)
		a[1] = utils.GetCdnURL(r.gCtx.Config().CdnURL, r.Emote.ID.Hex(), int8(i))

		result[i-1] = a
	}

	r.Emote.Links = result
	return r.Emote.Links
}

// Animated: whether or not the emote is animated
func (r *EmoteResolver) Animated() bool {
	return r.Emote.FrameCount > 1
}

// Owner: the user who owns the emote
func (r *EmoteResolver) Owner() (*UserResolver, error) {
	if r.Emote.OwnerID.IsZero() {
		return nil, nil
	}
	if ok := r.quota.Decrease(2, "Emote", "Owner"); !ok {
		return nil, nil
	}

	if r.Emote.Owner == nil {
		return nil, nil
	}

	return CreateUserResolver(r.gCtx, r.ctx, r.Emote.Owner, &r.Emote.Owner.ID, GenerateSelectedFieldMap(r.ctx).Children)
}

func (r *EmoteResolver) Channels(ctx context.Context, args struct {
	AfterID string
	Limit   *int32
}) ([]*UserResolver, error) {
	emote := r.Emote

	// Parse ID to go from
	// afterID, _ := primitive.ObjectIDFromHex(args.AfterID)

	// Define limit
	limit := int32(20)
	if args.Limit != nil {
		limit = *args.Limit
		if limit < 1 {
			limit = 1
		} else if limit > 250 {
			limit = 250
		}
	}

	// Pipeline
	pipeline := mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.M{
				"channel_emotes.id": emote.ID,
			},
		}},
		{{
			Key: "$lookup",
			Value: mongo.Lookup{
				From:         "roles",
				LocalField:   "role_ids",
				ForeignField: "_id",
				As:           "_role",
			},
		}},
		{{
			Key: "$sort",
			Value: bson.M{
				"_role.position": -1,
			},
		}},
		{{Key: "$limit", Value: limit}},
	}

	users := []*structures.User{}
	cur, err := r.gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, pipeline)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	if err = cur.All(ctx, &users); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	resolvers := []*UserResolver{}
	for _, usr := range users {
		resolver, err := CreateUserResolver(r.gCtx, ctx, usr, &usr.ID, r.fields)
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, resolver)
	}

	return resolvers, nil
}

func (r *EmoteResolver) ChannelCount(ctx context.Context) (int32, error) {
	emote := r.Emote

	count, err := r.gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).CountDocuments(ctx, bson.M{
		"channel_emotes.id": emote.ID,
	})
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return 0, err
	}

	return int32(count), nil
}
