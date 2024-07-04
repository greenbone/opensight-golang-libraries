package sorting

import (
	"fmt"
)

type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}

func NewSortingError(format string, value ...any) error {
	return &Error{
		Msg: fmt.Sprintf(format, value...),
	}
}
