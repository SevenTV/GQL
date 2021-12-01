package query

import (
	"context"
	"fmt"
	"time"

	"github.com/SevenTV/Common/aggregations"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/global"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BanResolver struct {
	ctx context.Context
	*structures.BanBuilder

	fields map[string]*SelectedField
	gCtx   global.Context
}

func CreateBanResolver(gCtx global.Context, ctx context.Context, ban *structures.Ban, banID *primitive.ObjectID, fields map[string]*SelectedField) (*BanResolver, error) {
	bb := structures.NewBanBuilder(ban)

	var pipeline mongo.Pipeline
	if ban == nil && banID == nil {
		return nil, fmt.Errorf("unresolvable")
	}
	if ban == nil {
		bb.Ban = &structures.Ban{}
		pipeline = mongo.Pipeline{
			{{
				Key:   "$match",
				Value: bson.M{"_id": banID},
			}},
		}
	} else {
		pipeline = mongo.Pipeline{
			{{
				Key:   "$replaceRoot",
				Value: bson.M{"newRoot": ban},
			}},
		}
	}

	// Relation: victim
	if _, ok := fields["victim"]; ok {
		pipeline = append(pipeline, aggregations.BanRelationVictim...)
	}

	// Relation: actor
	if _, ok := fields["actor"]; ok {
		pipeline = append(pipeline, aggregations.BanRelationActor...)
	}

	if bb.Ban.ID.IsZero() || len(pipeline) > 1 {
		cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameBans).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		cur.Next(ctx)
		cur.Close(ctx)
		if err = cur.Decode(bb.Ban); err != nil {
			logrus.WithError(err).Error("mongo")
			return nil, err
		}
	}

	return &BanResolver{
		ctx,
		bb,
		fields,
		gCtx,
	}, nil
}

func (r *BanResolver) ID() string {
	return r.Ban.ID.Hex()
}

func (r *BanResolver) Victim(ctx context.Context) (*UserResolver, error) {
	return CreateUserResolver(r.gCtx, ctx, r.Ban.Victim, &r.Ban.VictimID, r.fields)
}

func (r *BanResolver) Actor(ctx context.Context) (*UserResolver, error) {
	return CreateUserResolver(r.gCtx, ctx, r.Ban.Actor, &r.Ban.ActorID, r.fields)
}

func (r *BanResolver) Reason() string {
	return r.Ban.Reason
}

func (r *BanResolver) Effects() int32 {
	return int32(r.Ban.Effects)
}

func (r *BanResolver) ExpireAt() string {
	return r.Ban.ExpireAt.Format(time.RFC3339)
}
