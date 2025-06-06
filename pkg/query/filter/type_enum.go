// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package filter

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

const (
	// AggregateMetricSum is a AggregateMetric of type sum.
	AggregateMetricSum AggregateMetric = "sum"
	// AggregateMetricMin is a AggregateMetric of type min.
	AggregateMetricMin AggregateMetric = "min"
	// AggregateMetricMax is a AggregateMetric of type max.
	AggregateMetricMax AggregateMetric = "max"
	// AggregateMetricAvg is a AggregateMetric of type avg.
	AggregateMetricAvg AggregateMetric = "avg"
	// AggregateMetricValueCount is a AggregateMetric of type valueCount.
	AggregateMetricValueCount AggregateMetric = "valueCount"
)

var ErrInvalidAggregateMetric = fmt.Errorf("not a valid AggregateMetric, try [%s]", strings.Join(_AggregateMetricNames, ", "))

var _AggregateMetricNames = []string{
	string(AggregateMetricSum),
	string(AggregateMetricMin),
	string(AggregateMetricMax),
	string(AggregateMetricAvg),
	string(AggregateMetricValueCount),
}

// AggregateMetricNames returns a list of possible string values of AggregateMetric.
func AggregateMetricNames() []string {
	tmp := make([]string, len(_AggregateMetricNames))
	copy(tmp, _AggregateMetricNames)
	return tmp
}

// String implements the Stringer interface.
func (x AggregateMetric) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x AggregateMetric) IsValid() bool {
	_, err := ParseAggregateMetric(string(x))
	return err == nil
}

var _AggregateMetricValue = map[string]AggregateMetric{
	"sum":        AggregateMetricSum,
	"min":        AggregateMetricMin,
	"max":        AggregateMetricMax,
	"avg":        AggregateMetricAvg,
	"valueCount": AggregateMetricValueCount,
}

// ParseAggregateMetric attempts to convert a string to a AggregateMetric.
func ParseAggregateMetric(name string) (AggregateMetric, error) {
	if x, ok := _AggregateMetricValue[name]; ok {
		return x, nil
	}
	return AggregateMetric(""), fmt.Errorf("%s is %w", name, ErrInvalidAggregateMetric)
}

// MarshalText implements the text marshaller method.
func (x AggregateMetric) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *AggregateMetric) UnmarshalText(text []byte) error {
	tmp, err := ParseAggregateMetric(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

var errAggregateMetricNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *AggregateMetric) Scan(value interface{}) (err error) {
	if value == nil {
		*x = AggregateMetric("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParseAggregateMetric(v)
	case []byte:
		*x, err = ParseAggregateMetric(string(v))
	case AggregateMetric:
		*x = v
	case *AggregateMetric:
		if v == nil {
			return errAggregateMetricNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errAggregateMetricNilPtr
		}
		*x, err = ParseAggregateMetric(*v)
	default:
		return errors.New("invalid type for AggregateMetric")
	}

	return
}

// Value implements the driver Valuer interface.
func (x AggregateMetric) Value() (driver.Value, error) {
	return x.String(), nil
}

const (
	// CompareOperatorBeginsWith is a CompareOperator of type beginsWith.
	CompareOperatorBeginsWith CompareOperator = "beginsWith"
	// CompareOperatorDoesNotBeginWith is a CompareOperator of type doesNotBeginWith.
	CompareOperatorDoesNotBeginWith CompareOperator = "doesNotBeginWith"
	// CompareOperatorContains is a CompareOperator of type contains.
	CompareOperatorContains CompareOperator = "contains"
	// CompareOperatorDoesNotContain is a CompareOperator of type doesNotContain.
	CompareOperatorDoesNotContain CompareOperator = "doesNotContain"
	// CompareOperatorTextContains is a CompareOperator of type textContains.
	CompareOperatorTextContains CompareOperator = "textContains"
	// CompareOperatorIsNumberEqualTo is a CompareOperator of type isNumberEqualTo.
	CompareOperatorIsNumberEqualTo CompareOperator = "isNumberEqualTo"
	// CompareOperatorIsEqualTo is a CompareOperator of type isEqualTo.
	CompareOperatorIsEqualTo CompareOperator = "isEqualTo"
	// CompareOperatorIsIpEqualTo is a CompareOperator of type isIpEqualTo.
	CompareOperatorIsIpEqualTo CompareOperator = "isIpEqualTo"
	// CompareOperatorIsStringEqualTo is a CompareOperator of type isStringEqualTo.
	CompareOperatorIsStringEqualTo CompareOperator = "isStringEqualTo"
	// CompareOperatorIsStringCaseInsensitiveEqualTo is a CompareOperator of type isStringCaseInsensitiveEqualTo.
	CompareOperatorIsStringCaseInsensitiveEqualTo CompareOperator = "isStringCaseInsensitiveEqualTo"
	// CompareOperatorIsNotEqualTo is a CompareOperator of type isNotEqualTo.
	CompareOperatorIsNotEqualTo CompareOperator = "isNotEqualTo"
	// CompareOperatorIsNumberNotEqualTo is a CompareOperator of type isNumberNotEqualTo.
	CompareOperatorIsNumberNotEqualTo CompareOperator = "isNumberNotEqualTo"
	// CompareOperatorIsIpNotEqualTo is a CompareOperator of type isIpNotEqualTo.
	CompareOperatorIsIpNotEqualTo CompareOperator = "isIpNotEqualTo"
	// CompareOperatorIsStringNotEqualTo is a CompareOperator of type isStringNotEqualTo.
	CompareOperatorIsStringNotEqualTo CompareOperator = "isStringNotEqualTo"
	// CompareOperatorIsGreaterThan is a CompareOperator of type isGreaterThan.
	CompareOperatorIsGreaterThan CompareOperator = "isGreaterThan"
	// CompareOperatorIsGreaterThanOrEqualTo is a CompareOperator of type isGreaterThanOrEqualTo.
	CompareOperatorIsGreaterThanOrEqualTo CompareOperator = "isGreaterThanOrEqualTo"
	// CompareOperatorIsLessThan is a CompareOperator of type isLessThan.
	CompareOperatorIsLessThan CompareOperator = "isLessThan"
	// CompareOperatorIsLessThanOrEqualTo is a CompareOperator of type isLessThanOrEqualTo.
	CompareOperatorIsLessThanOrEqualTo CompareOperator = "isLessThanOrEqualTo"
	// CompareOperatorBeforeDate is a CompareOperator of type beforeDate.
	CompareOperatorBeforeDate CompareOperator = "beforeDate"
	// CompareOperatorAfterDate is a CompareOperator of type afterDate.
	CompareOperatorAfterDate CompareOperator = "afterDate"
	// CompareOperatorExists is a CompareOperator of type exists.
	CompareOperatorExists CompareOperator = "exists"
	// CompareOperatorIsEqualToRating is a CompareOperator of type isEqualToRating.
	CompareOperatorIsEqualToRating CompareOperator = "isEqualToRating"
	// CompareOperatorIsNotEqualToRating is a CompareOperator of type isNotEqualToRating.
	CompareOperatorIsNotEqualToRating CompareOperator = "isNotEqualToRating"
	// CompareOperatorIsGreaterThanRating is a CompareOperator of type isGreaterThanRating.
	CompareOperatorIsGreaterThanRating CompareOperator = "isGreaterThanRating"
	// CompareOperatorIsLessThanRating is a CompareOperator of type isLessThanRating.
	CompareOperatorIsLessThanRating CompareOperator = "isLessThanRating"
	// CompareOperatorIsGreaterThanOrEqualToRating is a CompareOperator of type isGreaterThanOrEqualToRating.
	CompareOperatorIsGreaterThanOrEqualToRating CompareOperator = "isGreaterThanOrEqualToRating"
	// CompareOperatorIsLessThanOrEqualToRating is a CompareOperator of type isLessThanOrEqualToRating.
	CompareOperatorIsLessThanOrEqualToRating CompareOperator = "isLessThanOrEqualToRating"
	// CompareOperatorBetweenDates is a CompareOperator of type betweenDates.
	CompareOperatorBetweenDates CompareOperator = "betweenDates"
)

var ErrInvalidCompareOperator = fmt.Errorf("not a valid CompareOperator, try [%s]", strings.Join(_CompareOperatorNames, ", "))

var _CompareOperatorNames = []string{
	string(CompareOperatorBeginsWith),
	string(CompareOperatorDoesNotBeginWith),
	string(CompareOperatorContains),
	string(CompareOperatorDoesNotContain),
	string(CompareOperatorTextContains),
	string(CompareOperatorIsNumberEqualTo),
	string(CompareOperatorIsEqualTo),
	string(CompareOperatorIsIpEqualTo),
	string(CompareOperatorIsStringEqualTo),
	string(CompareOperatorIsStringCaseInsensitiveEqualTo),
	string(CompareOperatorIsNotEqualTo),
	string(CompareOperatorIsNumberNotEqualTo),
	string(CompareOperatorIsIpNotEqualTo),
	string(CompareOperatorIsStringNotEqualTo),
	string(CompareOperatorIsGreaterThan),
	string(CompareOperatorIsGreaterThanOrEqualTo),
	string(CompareOperatorIsLessThan),
	string(CompareOperatorIsLessThanOrEqualTo),
	string(CompareOperatorBeforeDate),
	string(CompareOperatorAfterDate),
	string(CompareOperatorExists),
	string(CompareOperatorIsEqualToRating),
	string(CompareOperatorIsNotEqualToRating),
	string(CompareOperatorIsGreaterThanRating),
	string(CompareOperatorIsLessThanRating),
	string(CompareOperatorIsGreaterThanOrEqualToRating),
	string(CompareOperatorIsLessThanOrEqualToRating),
	string(CompareOperatorBetweenDates),
}

// CompareOperatorNames returns a list of possible string values of CompareOperator.
func CompareOperatorNames() []string {
	tmp := make([]string, len(_CompareOperatorNames))
	copy(tmp, _CompareOperatorNames)
	return tmp
}

// String implements the Stringer interface.
func (x CompareOperator) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x CompareOperator) IsValid() bool {
	_, err := ParseCompareOperator(string(x))
	return err == nil
}

var _CompareOperatorValue = map[string]CompareOperator{
	"beginsWith":                     CompareOperatorBeginsWith,
	"doesNotBeginWith":               CompareOperatorDoesNotBeginWith,
	"contains":                       CompareOperatorContains,
	"doesNotContain":                 CompareOperatorDoesNotContain,
	"textContains":                   CompareOperatorTextContains,
	"isNumberEqualTo":                CompareOperatorIsNumberEqualTo,
	"isEqualTo":                      CompareOperatorIsEqualTo,
	"isIpEqualTo":                    CompareOperatorIsIpEqualTo,
	"isStringEqualTo":                CompareOperatorIsStringEqualTo,
	"isStringCaseInsensitiveEqualTo": CompareOperatorIsStringCaseInsensitiveEqualTo,
	"isNotEqualTo":                   CompareOperatorIsNotEqualTo,
	"isNumberNotEqualTo":             CompareOperatorIsNumberNotEqualTo,
	"isIpNotEqualTo":                 CompareOperatorIsIpNotEqualTo,
	"isStringNotEqualTo":             CompareOperatorIsStringNotEqualTo,
	"isGreaterThan":                  CompareOperatorIsGreaterThan,
	"isGreaterThanOrEqualTo":         CompareOperatorIsGreaterThanOrEqualTo,
	"isLessThan":                     CompareOperatorIsLessThan,
	"isLessThanOrEqualTo":            CompareOperatorIsLessThanOrEqualTo,
	"beforeDate":                     CompareOperatorBeforeDate,
	"afterDate":                      CompareOperatorAfterDate,
	"exists":                         CompareOperatorExists,
	"isEqualToRating":                CompareOperatorIsEqualToRating,
	"isNotEqualToRating":             CompareOperatorIsNotEqualToRating,
	"isGreaterThanRating":            CompareOperatorIsGreaterThanRating,
	"isLessThanRating":               CompareOperatorIsLessThanRating,
	"isGreaterThanOrEqualToRating":   CompareOperatorIsGreaterThanOrEqualToRating,
	"isLessThanOrEqualToRating":      CompareOperatorIsLessThanOrEqualToRating,
	"betweenDates":                   CompareOperatorBetweenDates,
}

// ParseCompareOperator attempts to convert a string to a CompareOperator.
func ParseCompareOperator(name string) (CompareOperator, error) {
	if x, ok := _CompareOperatorValue[name]; ok {
		return x, nil
	}
	return CompareOperator(""), fmt.Errorf("%s is %w", name, ErrInvalidCompareOperator)
}

// MarshalText implements the text marshaller method.
func (x CompareOperator) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *CompareOperator) UnmarshalText(text []byte) error {
	tmp, err := ParseCompareOperator(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

var errCompareOperatorNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *CompareOperator) Scan(value interface{}) (err error) {
	if value == nil {
		*x = CompareOperator("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParseCompareOperator(v)
	case []byte:
		*x, err = ParseCompareOperator(string(v))
	case CompareOperator:
		*x = v
	case *CompareOperator:
		if v == nil {
			return errCompareOperatorNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errCompareOperatorNilPtr
		}
		*x, err = ParseCompareOperator(*v)
	default:
		return errors.New("invalid type for CompareOperator")
	}

	return
}

// Value implements the driver Valuer interface.
func (x CompareOperator) Value() (driver.Value, error) {
	return x.String(), nil
}

const (
	// ControlTypeBool is a ControlType of type bool.
	ControlTypeBool ControlType = "bool"
	// ControlTypeEnum is a ControlType of type enum.
	ControlTypeEnum ControlType = "enum"
	// ControlTypeFloat is a ControlType of type float.
	ControlTypeFloat ControlType = "float"
	// ControlTypeInteger is a ControlType of type integer.
	ControlTypeInteger ControlType = "integer"
	// ControlTypeString is a ControlType of type string.
	ControlTypeString ControlType = "string"
	// ControlTypeDateTime is a ControlType of type dateTime.
	ControlTypeDateTime ControlType = "dateTime"
	// ControlTypeUuid is a ControlType of type uuid.
	ControlTypeUuid ControlType = "uuid"
	// ControlTypeAutocomplete is a ControlType of type autocomplete.
	ControlTypeAutocomplete ControlType = "autocomplete"
)

var ErrInvalidControlType = fmt.Errorf("not a valid ControlType, try [%s]", strings.Join(_ControlTypeNames, ", "))

var _ControlTypeNames = []string{
	string(ControlTypeBool),
	string(ControlTypeEnum),
	string(ControlTypeFloat),
	string(ControlTypeInteger),
	string(ControlTypeString),
	string(ControlTypeDateTime),
	string(ControlTypeUuid),
	string(ControlTypeAutocomplete),
}

// ControlTypeNames returns a list of possible string values of ControlType.
func ControlTypeNames() []string {
	tmp := make([]string, len(_ControlTypeNames))
	copy(tmp, _ControlTypeNames)
	return tmp
}

// String implements the Stringer interface.
func (x ControlType) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x ControlType) IsValid() bool {
	_, err := ParseControlType(string(x))
	return err == nil
}

var _ControlTypeValue = map[string]ControlType{
	"bool":         ControlTypeBool,
	"enum":         ControlTypeEnum,
	"float":        ControlTypeFloat,
	"integer":      ControlTypeInteger,
	"string":       ControlTypeString,
	"dateTime":     ControlTypeDateTime,
	"uuid":         ControlTypeUuid,
	"autocomplete": ControlTypeAutocomplete,
}

// ParseControlType attempts to convert a string to a ControlType.
func ParseControlType(name string) (ControlType, error) {
	if x, ok := _ControlTypeValue[name]; ok {
		return x, nil
	}
	return ControlType(""), fmt.Errorf("%s is %w", name, ErrInvalidControlType)
}

// MarshalText implements the text marshaller method.
func (x ControlType) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *ControlType) UnmarshalText(text []byte) error {
	tmp, err := ParseControlType(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

var errControlTypeNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *ControlType) Scan(value interface{}) (err error) {
	if value == nil {
		*x = ControlType("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParseControlType(v)
	case []byte:
		*x, err = ParseControlType(string(v))
	case ControlType:
		*x = v
	case *ControlType:
		if v == nil {
			return errControlTypeNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errControlTypeNilPtr
		}
		*x, err = ParseControlType(*v)
	default:
		return errors.New("invalid type for ControlType")
	}

	return
}

// Value implements the driver Valuer interface.
func (x ControlType) Value() (driver.Value, error) {
	return x.String(), nil
}

const (
	// LogicOperatorAnd is a LogicOperator of type and.
	LogicOperatorAnd LogicOperator = "and"
	// LogicOperatorOr is a LogicOperator of type or.
	LogicOperatorOr LogicOperator = "or"
)

var ErrInvalidLogicOperator = fmt.Errorf("not a valid LogicOperator, try [%s]", strings.Join(_LogicOperatorNames, ", "))

var _LogicOperatorNames = []string{
	string(LogicOperatorAnd),
	string(LogicOperatorOr),
}

// LogicOperatorNames returns a list of possible string values of LogicOperator.
func LogicOperatorNames() []string {
	tmp := make([]string, len(_LogicOperatorNames))
	copy(tmp, _LogicOperatorNames)
	return tmp
}

// String implements the Stringer interface.
func (x LogicOperator) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x LogicOperator) IsValid() bool {
	_, err := ParseLogicOperator(string(x))
	return err == nil
}

var _LogicOperatorValue = map[string]LogicOperator{
	"and": LogicOperatorAnd,
	"or":  LogicOperatorOr,
}

// ParseLogicOperator attempts to convert a string to a LogicOperator.
func ParseLogicOperator(name string) (LogicOperator, error) {
	if x, ok := _LogicOperatorValue[name]; ok {
		return x, nil
	}
	return LogicOperator(""), fmt.Errorf("%s is %w", name, ErrInvalidLogicOperator)
}

// MarshalText implements the text marshaller method.
func (x LogicOperator) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *LogicOperator) UnmarshalText(text []byte) error {
	tmp, err := ParseLogicOperator(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

var errLogicOperatorNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *LogicOperator) Scan(value interface{}) (err error) {
	if value == nil {
		*x = LogicOperator("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParseLogicOperator(v)
	case []byte:
		*x, err = ParseLogicOperator(string(v))
	case LogicOperator:
		*x = v
	case *LogicOperator:
		if v == nil {
			return errLogicOperatorNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errLogicOperatorNilPtr
		}
		*x, err = ParseLogicOperator(*v)
	default:
		return errors.New("invalid type for LogicOperator")
	}

	return
}

// Value implements the driver Valuer interface.
func (x LogicOperator) Value() (driver.Value, error) {
	return x.String(), nil
}
