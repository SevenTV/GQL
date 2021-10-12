package instance

import (
	"context"

	"golang.org/x/oauth2"
)

type Youtube interface {
	GetYTGAuthURL() string
	GetYTGToken(ctx context.Context, code string) (*oauth2.Token, error)
}
