package errs

import "errors"

var (
	ErrInvalidParameters = errors.New("invalid request parameter")
)

var (
	ErrAgentIDRequired     = errors.New("agent ID is required")
	ErrAgentNameRequired   = errors.New("agent name is required")
	ErrChatModelIDRequired = errors.New("chat model ID is required")
)

var (
	ErrAgentNotFound = errors.New("agent not found")
)
