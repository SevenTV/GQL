package mutation

import (
	"context"

	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/graph/model"
	"github.com/SevenTV/GQL/src/server/v3/gql/types"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.MutationResolver {
	return &Resolver{r}
}

func (r *Resolver) SetChannelEmote(ctx context.Context, userID string, target model.UserEmoteInput, action model.ChannelEmoteListItemAction) (*model.User, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) SetUserRole(ctx context.Context, userID string, roleID string, action model.ChannelEmoteListItemAction) (*model.User, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) EditEmote(ctx context.Context, emoteID string, data model.EditEmoteInput) (*model.Emote, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) CreateRole(ctx context.Context, data model.CreateRoleInput) (*model.Role, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) EditRole(ctx context.Context, roleID string, data model.EditRoleInput) (*model.Role, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) DeleteRole(ctx context.Context, roleID string) (string, error) {
	// TODO
	return "", nil
}

func (r *Resolver) CreateReport(ctx context.Context, data model.CreateReportInput) (*model.Report, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) EditReport(ctx context.Context, reportID string, data model.EditReportInput) (*model.Report, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) CreateBan(ctx context.Context, victimID string, reason string, effects int, expireAt *string, anonymous *bool) (*model.Ban, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) EditBan(ctx context.Context, banID string, reason *string, effects *int, expireAt *string) (*model.Ban, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) ReadMessages(ctx context.Context, messageIds []string, read *bool) (*int, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) SendInboxMessage(ctx context.Context, recipients []string, subject string, content string, important *bool, anonymous *bool) (*model.Message, error) {
	// TODO
	return nil, nil
}
