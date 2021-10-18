package query

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
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
	eb := structures.EmoteBuilder{Emote: emote}

	if eb.Emote == nil && emoteID == nil {
		return nil, fmt.Errorf("Unresolvable")
	}
	if eb.Emote == nil {
		if _, err := eb.FetchByID(ctx, gCtx.Inst().Mongo, *emoteID); err != nil {
			return nil, err
		}
	}

	return &EmoteResolver{
		ctx:          ctx,
		quota:        ctx.Value(helpers.QuotaKey).(*helpers.Quota),
		EmoteBuilder: &eb,
		fields:       fields,
		gCtx:         gCtx,
	}, nil
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
func (r *EmoteResolver) Visibility() int32 {
	return r.Emote.Visibility
}

// Status: emote status
func (r *EmoteResolver) Status() int32 {
	return r.Emote.Status
}

// Tags: emote search tags
func (r *EmoteResolver) Tags() []string {
	return r.Emote.Tags
}

// URLs: resolves a list of cdn urls for the emote
func (r *EmoteResolver) URLs() [][]string {
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

	r.Emote.URLs = result
	return r.Emote.URLs
}

// Width: the emote's image width
func (r *EmoteResolver) Width() []int32 {
	return r.Emote.Width
}

// Height: the emote's image height
func (r *EmoteResolver) Height() []int32 {
	return r.Emote.Height
}

// Animated: whether or not the emote is animated
func (r *EmoteResolver) Animated() bool {
	return r.Emote.Animated
}

// Owner: the user who owns the emote
func (r *EmoteResolver) Owner() (*UserResolver, error) {
	if ok := r.quota.Decrease(2, "Emote", "Owner"); !ok {
		return nil, nil
	}

	if r.Emote.Owner == nil {
		return nil, nil
	}

	return CreateUserResolver(r.gCtx, r.ctx, r.Emote.Owner, &r.Emote.Owner.ID, GenerateSelectedFieldMap(r.ctx).Children)
}
