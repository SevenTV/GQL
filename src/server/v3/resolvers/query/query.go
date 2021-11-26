package query

import (
	"context"

	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/global"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/selection"
)

type Resolver struct {
	Ctx global.Context
}

func Fields(ctx context.Context) []*selection.SelectedField {
	return graphql.SelectedFieldsFromContext(ctx)
}

type SelectedField struct {
	Name     string
	Children map[string]*SelectedField
}

func GenerateSelectedFieldMap(ctx context.Context) *SelectedField {
	var loop func(fields []*selection.SelectedField) map[string]*SelectedField
	loop = func(fields []*selection.SelectedField) map[string]*SelectedField {
		m := map[string]*SelectedField{}
		for _, f := range fields {
			children := loop(f.SelectedFields)
			m[f.Name] = &SelectedField{
				Name:     f.Name,
				Children: children,
			}
		}
		return m
	}
	children := loop(Fields(ctx))
	return &SelectedField{
		Name:     "query",
		Children: children,
	}
}

type Sort struct {
	Value string    `json:"value"`
	Order SortOrder `json:"order"`
}

type SortOrder string

var (
	SortOrderAscending  SortOrder = "ASCENDING"
	SortOrderDescending SortOrder = "DESCENDING"
)

var sortOrderMap = map[string]int32{
	string(SortOrderDescending): 1,
	string(SortOrderAscending):  -1,
}

// GetDefaultRoles: Get a list of default roles
func GetDefaultRoles(gCtx global.Context, ctx context.Context) []*structures.Role {
	return nil
}
