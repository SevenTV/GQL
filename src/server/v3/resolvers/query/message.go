package query

import (
	"context"
	"fmt"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageResolver struct {
	ctx context.Context
	*structures.MessageBuilder

	fields map[string]*SelectedField
	gCtx   global.Context
}

func CreateMessageResolver(gCtx global.Context, ctx context.Context, msg *structures.Message, msgID *primitive.ObjectID, fields map[string]*SelectedField) (*MessageResolver, error) {
	mb := structures.NewMessageBuilder(msg)

	var pipeline mongo.Pipeline
	if msg == nil && msgID == nil {
		return nil, fmt.Errorf("unresolvable")
	}
	if msg == nil {
		mb.Message = &structures.Message{}
		pipeline = mongo.Pipeline{
			{{
				Key:   "$match",
				Value: bson.M{"_id": msgID},
			}},
		}
	} else {
		pipeline = mongo.Pipeline{
			{{
				Key:   "$replaceRoot",
				Value: bson.M{"newRoot": msg},
			}},
		}
	}

	if mb.Message.ID.IsZero() || len(pipeline) > 1 {
		cur, err := gCtx.Inst().Mongo.Collection(mongo.CollectionNameMessages).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		cur.Next(ctx)
		cur.Close(ctx)
		if err = cur.Decode(mb.Message); err != nil {
			logrus.WithError(err).Error("mongo")
			return nil, err
		}
	}

	return &MessageResolver{
		ctx:            ctx,
		MessageBuilder: mb,
		fields:         fields,
		gCtx:           gCtx,
	}, nil
}

func (r *MessageResolver) ID() string {
	return r.Message.ID.Hex()
}

func (r *MessageResolver) Kind() structures.MessageKind {
	return r.Message.Kind
}

func (r *MessageResolver) CreatedAt() string {
	return r.Message.CreatedAt.Format(time.RFC3339)
}

func (r *MessageResolver) Author(ctx context.Context) (*UserResolver, error) {
	if r.Message.Author == nil {
		return nil, nil
	}
	// Omit the author if message is anonymous
	// (Unless the user is privileged)
	if r.Message.Anonymous {
		actor := ctx.Value(helpers.UserKey).(*structures.User)
		if actor == nil {
			return nil, nil
		}
		if !actor.HasPermission(structures.RolePermissionBypassPrivacy) {
			return nil, nil
		}
	}

	return CreateUserResolver(r.gCtx, ctx, r.Message.Author, &r.Message.AuthorID, r.fields)
}

func (r *MessageResolver) Data() string {
	return string(r.Message.Data)
}
