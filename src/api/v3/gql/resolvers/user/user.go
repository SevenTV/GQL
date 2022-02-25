package user

import (
	"context"
	"sort"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.UserResolver {
	return &Resolver{r}
}

// Roles resolves the roles of a user
func (r *Resolver) Roles(ctx context.Context, obj *model.User) ([]*model.Role, error) {
	m := make(map[primitive.ObjectID]*model.Role)
	defaults, _ := r.Ctx.Inst().Query.Roles(ctx, bson.M{"default": true})

	for _, rol := range obj.Roles {
		m[rol.ID] = rol
	}
	for _, rol := range defaults {
		if _, ok := m[rol.ID]; ok {
			continue
		}
		m[rol.ID] = helpers.RoleStructureToModel(r.Ctx, rol)
	}

	result := make([]*model.Role, len(m))
	i := 0
	for _, rol := range m {
		result[i] = rol
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		a := result[i]
		b := result[j]
		return a.Position > b.Position
	})
	return result, nil
}

func (r *Resolver) EmoteSets(ctx context.Context, obj *model.User) ([]*model.EmoteSet, error) {
	return loaders.For(ctx).EmoteSetByUserID.Load(obj.ID)
}

// Connections lists the users' connections
func (r *Resolver) Connections(ctx context.Context, obj *model.User, platforms []model.ConnectionPlatform) ([]*model.UserConnection, error) {
	result := []*model.UserConnection{}
	for _, conn := range obj.Connections {
		ok := false
		if len(platforms) > 0 {
			for _, p := range platforms {
				if conn.Platform == p {
					ok = true
					break
				}
			}
		} else {
			ok = true
		}
		if ok {
			result = append(result, conn)
		}
	}
	return result, nil
}

// Editors returns a users' list of editors
func (r *Resolver) Editors(ctx context.Context, obj *model.User) ([]*model.UserEditor, error) {
	ids := make([]primitive.ObjectID, len(obj.Editors))
	for i, v := range obj.Editors {
		ids[i] = v.ID
	}
	users, errs := loaders.For(ctx).UserByID.LoadAll(ids)

	result := []*model.UserEditor{}
	for _, e := range obj.Editors {
		for _, u := range users {
			if e.ID == u.ID {
				e.User = helpers.UserStructureToPartialModel(r.Ctx, u)
				result = append(result, e)
				break
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		a := result[i]
		b := result[j]

		return a.AddedAt.After(b.AddedAt)
	})
	return result, multierror.Append(nil, errs...).ErrorOrNil()
}

func (r *Resolver) EditorOf(ctx context.Context, obj *model.User) ([]*model.UserEditor, error) {
	result := []*model.UserEditor{}
	editables, err := r.Ctx.Inst().Query.UserEditorOf(ctx, obj.ID)
	if err == nil {
		for _, ed := range editables {
			if ed.HasPermission(structures.UserEditorPermissionModifyEmotes) {
				result = append(result, helpers.UserEditorStructureToModel(r.Ctx, ed))
			}
		}
	}

	return result, err
}

func (r *Resolver) OwnedEmotes(ctx context.Context, obj *model.User) ([]*model.Emote, error) {
	emotes := []*structures.Emote{}
	errs := []error{}
	cur, err := r.Ctx.Inst().Mongo.Collection(mongo.CollectionNameEmotes).Find(ctx, bson.M{
		"owner_id": obj.ID,
	})
	if err == nil {
		if err = cur.All(ctx, &emotes); err != nil {
			logrus.WithError(err).Error("mongo, failed to retrieve user's owned emotes")
			errs = append(errs, errors.ErrUnknownEmote())
		}
	}
	result := make([]*model.Emote, len(emotes))
	for i, e := range emotes {
		result[i] = helpers.EmoteStructureToModel(r.Ctx, e)
	}
	return result, multierror.Append(nil, errs...).ErrorOrNil()
}

func (r *Resolver) InboxUnreadCount(ctx context.Context, obj *model.User) (int, error) {
	// TODO
	return 0, nil
}

func (r *Resolver) Reports(ctx context.Context, obj *model.User) ([]*model.Report, error) {
	return loaders.For(ctx).ReportsByUserID.Load(obj.ID)
}
