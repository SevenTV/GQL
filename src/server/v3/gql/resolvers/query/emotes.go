package query

import (
	"context"
)

const EMOTES_QUERY_LIMIT int32 = 150

type EmoteListResponse struct {
	Total  int32    `json:"total"`
	Emotes []string `json:"emotes"`
}

func (r *Resolver) Emotes(ctx context.Context, args struct {
	Query    string
	Limit    *int32
	AfterID  *string
	BeforeID *string
}) (*EmoteListResponse, error) {
	// Define limit (how many emotes can be returned in a single query)
	limit := int32(20)
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit > EMOTES_QUERY_LIMIT {
		limit = EMOTES_QUERY_LIMIT
	}

	fields := GenerateSelectedFieldMap(ctx)
	EmoteResolver(ctx, nil, nil, fields.Children)

	return nil, nil
}
