package emote

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
)

type ResolverVersion struct {
	types.Resolver
}

func NewVersion(r types.Resolver) generated.EmoteVersionResolver {
	return &ResolverVersion{r}
}

func (r *ResolverVersion) Images(ctx context.Context, obj *model.EmoteVersion, format []model.ImageFormat) ([]*model.Image, error) {
	result := []*model.Image{}
	for _, im := range obj.Images {
		ok := len(format) == 0
		if !ok {
			for _, f := range format {
				if im.Format == f {
					result = append(result, im)
				}
			}
			continue
		}

		result = append(result, im)
	}

	return result, nil
}
