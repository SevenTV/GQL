package resolvers

import (
	"github.com/SevenTV/GQL/graph/generated"

	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/ban"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/emote"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/emoteset"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/mutation"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/query"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/report"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/role"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/subscription"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers/user"
	user_editor "github.com/SevenTV/GQL/src/api/v3/gql/resolvers/user-editor"
	user_emote "github.com/SevenTV/GQL/src/api/v3/gql/resolvers/user-emote"

	"github.com/SevenTV/GQL/src/api/v3/gql/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.ResolverRoot {
	return &Resolver{
		Resolver: r,
	}
}

func (r *Resolver) Ban() generated.BanResolver {
	return ban.New(r.Resolver)
}

func (r *Resolver) Emote() generated.EmoteResolver {
	return emote.New(r.Resolver)
}

func (r *Resolver) EmotePartial() generated.EmotePartialResolver {
	return emote.NewPartial(r.Resolver)
}

func (r *Resolver) EmoteVersion() generated.EmoteVersionResolver {
	return emote.NewVersion(r.Resolver)
}

func (r *Resolver) Mutation() generated.MutationResolver {
	return mutation.New(r.Resolver)
}

func (r *Resolver) Query() generated.QueryResolver {
	return query.New(r.Resolver)
}

func (r *Resolver) Subscription() generated.SubscriptionResolver {
	return subscription.New(r.Resolver)
}

func (r *Resolver) Report() generated.ReportResolver {
	return report.New(r.Resolver)
}

func (r *Resolver) Role() generated.RoleResolver {
	return role.New(r.Resolver)
}

func (r *Resolver) User() generated.UserResolver {
	return user.New(r.Resolver)
}

func (r *Resolver) UserPartial() generated.UserPartialResolver {
	return user.NewPartial(r.Resolver)
}

func (r *Resolver) UserOps() generated.UserOpsResolver {
	return user.NewOps(r.Resolver)
}

func (r *Resolver) UserEditor() generated.UserEditorResolver {
	return user_editor.New(r.Resolver)
}

func (r *Resolver) UserEmote() generated.UserEmoteResolver {
	return user_emote.New(r.Resolver)
}

func (r *Resolver) EmoteSet() generated.EmoteSetResolver {
	return emoteset.New(r.Resolver)
}

func (r *Resolver) EmoteSetOps() generated.EmoteSetOpsResolver {
	return emoteset.NewOps(r.Resolver)
}
