package helpers

import "fmt"

type ErrorGQL error

var (
	ErrAccessDenied ErrorGQL = fmt.Errorf("access denied")
)
