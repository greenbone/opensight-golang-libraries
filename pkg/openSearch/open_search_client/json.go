// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package open_search_client

import (
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
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
			iter.Error = errors.WithStack(err)
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
		return errors.WithStack(err)
	}

	return nil
}

func UnmarshalWithoutValidation(data []byte, v any) error {
	reportUpdatedEvent := v
	if err := jsoniter.Unmarshal(data, &reportUpdatedEvent); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func Marshal(v any) ([]byte, error) {
	marshal, err := jsoniter.Marshal(v)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return marshal, nil
}

func StructFields(value any) []string {
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

func ParseUnknownFields(
	bytes []byte,
	structValue any,
	callback func(iter *jsoniter.Iterator, fieldName string, callbackExtra any),
	callbackExtra any,
) {
	structFields := StructFields(structValue)
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
