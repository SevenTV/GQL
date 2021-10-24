package helpers

import "fmt"

type ErrorGQL error

var (
	ErrAccessDenied ErrorGQL = fmt.Errorf("access denied")
	ErrUnknownEmote ErrorGQL = fmt.Errorf("unknown emote")
)
