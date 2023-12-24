package analyzing

import "errors"

var (
	ErrEmptyTokens        = errors.New("empty tokens")
	ErrUnknownCommand     = errors.New("unknown command")
	ErrNotEnoughArguments = errors.New("not enough arguments")
	ErrTooMuchArguments   = errors.New("too much arguments")
)
