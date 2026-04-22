package errs

import "errors"

var (
	ErrInvalidParameters = errors.New("invalid request parameter")
)
