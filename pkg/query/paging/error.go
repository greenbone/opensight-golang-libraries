package paging

import (
	"fmt"

	"github.com/pkg/errors"
)

type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}

func NewPagingError(format string, value ...any) error {
	return errors.WithStack(&Error{
		Msg: fmt.Sprintf(format, value...),
	})
}
