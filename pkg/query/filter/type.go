// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package filter

//go:generate sh -c "test type_enum.go -nt $GOFILE && exit 0; go run \"github.com/abice/go-enum\" --marshal --sql --names"

/*
ControlType ENUM(

	bool
	enum
	float
	integer
	string
	dateTime
	uuid
	autocomplete

)
*/
type ControlType string

func (x ControlType) Cmp(other ControlType) int {
	return cmpStringBackedEnum(_ControlTypeNames, string(x), string(other))
}

/*
LogicOperator ENUM(

	and
	or

)
*/
type LogicOperator string

func (x LogicOperator) Cmp(other LogicOperator) int {
	return cmpStringBackedEnum(_LogicOperatorNames, string(x), string(other))
}

/*
CompareOperator ENUM(

	beginsWith
	doesNotBeginWith

	contains
	doesNotContain

	textContains

	isNumberEqualTo
	isNumberNotEqualTo

	isEqualTo
	isNotEqualTo

	isIpEqualTo
	isIpNotEqualTo

	isStringEqualTo
	isStringNotEqualTo
	isStringCaseInsensitiveEqualTo

	isGreaterThan
	isGreaterThanOrEqualTo

	isLessThan
	isLessThanOrEqualTo

	beforeDate
	afterDate
	betweenDates

	exists
	mustNotExists

	isEqualToRating
	isNotEqualToRating

	isLessThanRating
	isLessThanOrEqualToRating

	isGreaterThanRating
	isGreaterThanOrEqualToRating

)
*/
type CompareOperator string

func (x CompareOperator) Cmp(other CompareOperator) int {
	return cmpStringBackedEnum(_CompareOperatorNames, string(x), string(other))
}

/*
AggregateMetric ENUM(

	sum
	min
	max
	avg

	valueCount

)
*/
type AggregateMetric string

func (x AggregateMetric) Cmp(other AggregateMetric) int {
	return cmpStringBackedEnum(_AggregateMetricNames, string(x), string(other))
}
