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

/*
LogicOperator ENUM(

	and
	or

)
*/
type LogicOperator string

/*
CompareOperator ENUM(

				beginsWith
				doesNotBeginWith
				contains
				doesNotContain

				textContains

				isNumberEqualTo
				isEqualTo
				isIpEqualTo
				isStringEqualTo
				isStringCaseInsensitiveEqualTo

				isNotEqualTo
				isNumberNotEqualTo
				isIpNotEqualTo
				isStringNotEqualTo

				isGreaterThan
				isGreaterThanOrEqualTo
				isLessThan
				isLessThanOrEqualTo
				beforeDate
				afterDate

				exists

				isEqualToRating
	            isNotEqualToRating
			    isGreaterThanRating
		        isLessThanRating
				isGreaterThanOrEqualToRating
		        isLessThanOrEqualToRating

				betweenDates

)
*/
type CompareOperator string

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
