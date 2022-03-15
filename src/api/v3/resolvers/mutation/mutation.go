package mutation

import (
	"context"

	"github.com/SevenTV/GQL/graph/v3/generated"
	"github.com/SevenTV/GQL/graph/v3/model"
	"github.com/SevenTV/GQL/src/api/v3/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resolver struct {
	types.Resolver
}

func New(r types.Resolver) generated.MutationResolver {
	return &Resolver{r}
}

func (r *Resolver) SetUserRole(ctx context.Context, userID primitive.ObjectID, roleID primitive.ObjectID, action model.ListItemAction) (*model.User, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) CreateRole(ctx context.Context, data model.CreateRoleInput) (*model.Role, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) EditRole(ctx context.Context, roleID primitive.ObjectID, data model.EditRoleInput) (*model.Role, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) DeleteRole(ctx context.Context, roleID primitive.ObjectID) (string, error) {
	// TODO
	return "", nil
}

func (r *Resolver) CreateReport(ctx context.Context, data model.CreateReportInput) (*model.Report, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) EditReport(ctx context.Context, reportID primitive.ObjectID, data model.EditReportInput) (*model.Report, error) {
	// primitive.ObjectID
	return nil, nil
}

func (r *Resolver) CreateBan(ctx context.Context, victimID primitive.ObjectID, reason string, effects int, expireAt *string, anonymous *bool) (*model.Ban, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) EditBan(ctx context.Context, banID primitive.ObjectID, reason *string, effects *int, expireAt *string) (*model.Ban, error) {
	// TODO
	return nil, nil
}

func (r *Resolver) SendInboxMessage(ctx context.Context, recipients []string, subject string, content string, important *bool, anonymous *bool) (*model.InboxMessage, error) {
	// TODO
	return nil, nil
}
