package errs

import "errors"

var (
	ErrInvalidParameters = errors.New("invalid request parameter")
)

var (
	ErrAgentIDRequired   = errors.New("agent ID is required")
	ErrSessionIDRequired = errors.New("session ID is required")
)
