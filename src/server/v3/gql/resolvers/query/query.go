package query

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/selection"
)

type Resolver struct{}

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
