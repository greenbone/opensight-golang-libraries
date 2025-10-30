package filter

import (
	"cmp"
	"slices"
	"strings"
)

func cmpAny[T any](x, y T) int {
	if v, ok := any(x).(interface{ Cmp(T) int }); ok {
		return v.Cmp(y)
	}
	return 0 // unorderable value
}

func cmpStringBackedEnum(values []string, x, y string) int {
	i, j := slices.Index(values, x), slices.Index(values, y)
	switch {
	case i >= 0 && j >= 0: // both x and y are valid, order based on declaration order
		return cmp.Compare(i, j)
	case i >= 0: // y has invalid value, x goes before y
		return -1
	case j >= 0: // x has invalid value, y goes before x
		return 1
	default: // both x and y have invalid values, use strings values to determine order
		return strings.Compare(x, y)
	}
}
