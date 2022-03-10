package query

import (
	"context"
	"sort"
	"strings"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/query"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/helpers"
	"github.com/SevenTV/GQL/src/api/v3/loaders"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const EMOTES_QUERY_LIMIT = 300

func (r *Resolver) Emote(ctx context.Context, id primitive.ObjectID) (*model.Emote, error) {
	emote, err := loaders.For(ctx).EmoteByID.Load(id)
	if emote == nil || emote.ID == structures.DeletedEmote.ID {
		return nil, errors.ErrUnknownEmote()
	}
	return emote, err
}

func (r *Resolver) Emotes(ctx context.Context, queryValue string, pageArg *int, limitArg *int, filterArg *model.EmoteSearchFilter, sortArg *model.Sort) (*model.EmoteSearchResult, error) {
	// Define limit (how many emotes can be returned in a single query)
	limit := 20
	if limitArg != nil {
		limit = *limitArg
	}
	if limit > EMOTES_QUERY_LIMIT {
		limit = EMOTES_QUERY_LIMIT
	}

	// Define default filter
	filter := filterArg
	if filter == nil {
		filter = &model.EmoteSearchFilter{
			CaseSensitive: utils.BoolPointer(false),
			ExactMatch:    utils.BoolPointer(false),
		}
	} else {
		filter = filterArg
	}

	// Define the query string
	queryValue = strings.Trim(queryValue, " ")

	// Retrieve pagination values
	page := 1
	if pageArg != nil {
		page = *pageArg
	}
	if page < 1 {
		page = 1
	}

	// Retrieve sorting options
	sortopt := &model.Sort{
		Value: "popularity",
		Order: model.SortOrderAscending,
	}
	if sortArg != nil {
		sortopt = sortArg
	}

	// Define sorting
	// (will be ignored in the case of exact search)
	order, validOrder := sortOrderMap[string(sortopt.Order)]
	field, validField := sortFieldMap[sortopt.Value]
	sortMap := bson.M{}
	if validField && validOrder {
		sortMap = bson.M{field: order}
	}

	// Run query
	result, totalCount, err := r.Ctx.Inst().Query.SearchEmotes(ctx, query.SearchEmotesOptions{
		Query: queryValue,
		Page:  page,
		Limit: limit,
		Sort:  sortMap,
		Filter: &query.SearchEmotesFilter{
			CaseSensitive: filter.CaseSensitive,
			ExactMatch:    filter.ExactMatch,
			IgnoreTags:    filter.IgnoreTags,
		},
	})
	if err != nil {
		return nil, err
	}

	models := make([]*model.Emote, len(result))
	for i, e := range result {
		// Sort by version timestamp
		sort.Slice(e.Versions, func(i, j int) bool {
			a := e.Versions[i]
			b := e.Versions[j]

			return b.Timestamp.Before(a.Timestamp)
		})

		// Bring forward the latest version
		if len(e.Versions) > 0 {
			e.ID = e.Versions[0].ID
		}
		models[i] = helpers.EmoteStructureToModel(r.Ctx, e)
	}

	return &model.EmoteSearchResult{
		Count: totalCount,
		Items: models,
	}, nil
}

var sortFieldMap = map[string]string{
	"age":        "_id",
	"popularity": "versions.state.channel_count",
}