// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"reflect"
	"strings"
	"time"
	"unsafe"

	"errors"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

func InitializeJson(timeFormats []string) {
	jsoniter.RegisterTypeDecoderFunc("time.Time", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		var err error
		var parsedTime time.Time
		readString := iter.ReadString()
		for _, timeFormat := range timeFormats {
			parsedTime, err = time.ParseInLocation(timeFormat, readString, time.UTC)
			if err == nil {
				*((*time.Time)(ptr)) = parsedTime
				return
			}
		}

		if err != nil {
			iter.Error = err
			return
		}
	})
}

func newJsonValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return validate
}

// Unmarshal unmarshalls data into v. It returns an error if the data is invalid.
func Unmarshal(data []byte, v any) error {
	err := UnmarshalWithoutValidation(data, v)
	if err != nil {
		return err
	}

	return validateStruct(v)
}

func validateStruct(v any) error {
	validate := newJsonValidator()

	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors

	if errors.As(err, &validationErrors) {
		return err
	}

	return nil
}

// UnmarshalWithoutValidation unmarshalls data into v. It returns an error if the data can not be parsed.
func UnmarshalWithoutValidation(data []byte, v any) error {
	reportUpdatedEvent := v
	if err := jsoniter.Unmarshal(data, &reportUpdatedEvent); err != nil {
		return err
	}
	return nil
}

func structFields(value any) []string {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}

	fields := []string{}
	for fieldIndex := 0; fieldIndex < typ.NumField(); fieldIndex++ {
		field := typ.Field(fieldIndex)
		if jsonTag, ok := field.Tag.Lookup("json"); ok {
			jsonField, _, _ := strings.Cut(jsonTag, ",")
			fields = append(fields, jsonField)
		}
	}
	return fields
}

func parseUnknownFields(
	bytes []byte,
	structValue any,
	callback func(iter *jsoniter.Iterator, fieldName string, callbackExtra any),
	callbackExtra any,
) {
	structFields := structFields(structValue)
	iterator := jsoniter.ParseBytes(jsoniter.ConfigDefault, bytes)

	var fieldName string
	for fieldName = iterator.ReadObject(); fieldName != ""; fieldName = iterator.ReadObject() {
		isStructField := false
		for _, structField := range structFields {
			if fieldName == structField {
				isStructField = true
				break
			}
		}

		if isStructField {
			iterator.Skip()
			continue
		}

		callback(iterator, fieldName, callbackExtra)
	}
}
