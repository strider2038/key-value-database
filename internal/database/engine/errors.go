package engine

import "fmt"

type BadRequestError struct {
	err error
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("bad request: %s", e.err)
}

func (e *BadRequestError) Unwrap() error {
	return e.err
}
