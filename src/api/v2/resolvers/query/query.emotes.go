package query

import (
	"context"
	"sort"
	"strconv"

	"github.com/SevenTV/Common/structures/v3/query"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/v2/model"
	"github.com/SevenTV/GQL/src/api/v2/helpers"
	"github.com/SevenTV/GQL/src/api/v2/loaders"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *Resolver) Emote(ctx context.Context, id string) (*model.Emote, error) {
	return loaders.For(ctx).EmoteByID.Load(id)
}

func (r *Resolver) SearchEmotes(
	ctx context.Context,
	queryArg string,
	limitArg *int,
	pageArg *int,
	pageSizeArg *int,
	submittedBy *string,
	globalState *string,
	sortByArg *string,
	sortOrderArg *int,
	channel *string,
	filter *model.EmoteFilter,
) ([]*model.Emote, error) {
	// Define page
	page := 1
	if pageArg != nil && *pageArg > 1 {
		page = *pageArg
	}
	// Define limit
	// This is how many emotes can be searched in one request at most
	limit := 20
	if limitArg != nil {
		limit = *limitArg
	}
	if limit > query.EMOTES_QUERY_LIMIT {
		limit = query.EMOTES_QUERY_LIMIT
	}

	// Define sorting
	if sortByArg == nil {
		sortByArg = utils.StringPointer("popularity")
	}
	if sortOrderArg == nil {
		sortOrderArg = utils.IntPointer(0)
	}
	sortField, validField := sortFieldMap[*sortByArg]
	sortOrder, validOrder := sortOrderMap[*sortOrderArg]
	sortMap := bson.M{}
	if validField && validOrder {
		sortMap = bson.M{sortField: sortOrder}
	}

	result, totalCount, err := r.Ctx.Inst().Query.SearchEmotes(ctx, query.SearchEmotesOptions{
		Query: queryArg,
		Page:  page,
		Limit: limit,
		Sort:  sortMap,
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

	rctx := ctx.Value(helpers.RequestCtxKey).(*fasthttp.RequestCtx)
	if rctx != nil {
		rctx.Response.Header.Set("X-Collection-Size", strconv.Itoa(totalCount))
	}
	return models, nil
}

var sortFieldMap = map[string]string{
	"age":        "_id",
	"popularity": "versions.state.channel_count",
}

var sortOrderMap = map[int]int{
	1: 1,
	0: -1,
}
