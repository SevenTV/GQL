package user

import (
	"context"
	"sort"

	"github.com/SevenTV/GQL/graph/v3/generated"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/helpers"
	"github.com/SevenTV/GQL/src/api/v3/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResolverPartial struct {
	types.Resolver
}

func NewPartial(r types.Resolver) generated.UserPartialResolver {
	return &ResolverPartial{r}
}

func (r *ResolverPartial) Roles(ctx context.Context, obj *model.UserPartial) ([]*model.Role, error) {
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

	result := make([]*model.Role, 0, len(m))
	for _, rol := range m {
		result = append(result, rol)
	}
	sort.Slice(result, func(i, j int) bool {
		a := result[i]
		b := result[j]
		return a.Position > b.Position
	})
	return result, nil
}