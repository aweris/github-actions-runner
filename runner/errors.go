package runner

import "errors"

var (
	ErrMissingParameter = errors.New("missing parameter")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrGHRequestFailed  = errors.New("github request failed")
)
