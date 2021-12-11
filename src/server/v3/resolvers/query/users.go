package query

import "context"

func (r *Resolver) Users(ctx context.Context, args struct {
	Query string
}) ([]*UserResolver, error) {
	return nil, nil
}
