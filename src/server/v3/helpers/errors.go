package helpers

import "fmt"

type ErrorGQL error

var (
	ErrUnauthorized        ErrorGQL = fmt.Errorf("unauthorized")
	ErrAccessDenied        ErrorGQL = fmt.Errorf("access denied")
	ErrUnknownEmote        ErrorGQL = fmt.Errorf("unknown emote")
	ErrUnknownUser         ErrorGQL = fmt.Errorf("unknown user")
	ErrUnknownRole         ErrorGQL = fmt.Errorf("unknown role")
	ErrBadObjectID         ErrorGQL = fmt.Errorf("bad object id")
	ErrInternalServerError ErrorGQL = fmt.Errorf("internal server error")
)
