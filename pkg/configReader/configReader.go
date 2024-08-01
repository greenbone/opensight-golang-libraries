package configReader

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// ReadEnvVarsIntoStruct reads environment variables into a given struct
func ReadEnvVarsIntoStruct(s any) (any, error) {
	structValue := reflect.ValueOf(s).Elem()
	structType := structValue.Type()

	viper.AutomaticEnv()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := field.Type()
		fieldTag := structType.Field(i).Tag.Get("viperEnv")
		defaultTag := structType.Field(i).Tag.Get("default")

		envValue := viper.GetString(fieldTag)
		if envValue == "" && fieldType.Kind() != reflect.Struct && fieldType.Kind() != reflect.Ptr {
			var ok bool
			envValue, ok = os.LookupEnv(fieldTag)
			if !ok {
				if defaultTag != "" {
					envValue = defaultTag // set the environment variable to the default value if it's not set and a default value is specified
				} else {
					continue // ignore the field if the environment variable is not set and there is no default value
				}
			}
		}

		switch fieldType.Kind() {
		case reflect.String:
			field.SetString(envValue)
		case reflect.Bool:
			boolValue := viper.GetBool(fieldTag)
			field.SetBool(boolValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fieldType == reflect.TypeOf(time.Duration(0)) {
				durationValue := viper.GetDuration(fieldTag)
				if durationValue == 0 && envValue != "" {
					envDur, _ := time.ParseDuration(envValue)
					durationValue = time.Duration(envDur)
				}
				field.Set(reflect.ValueOf(durationValue))
			} else {
				intValue := viper.GetInt64(fieldTag)
				if intValue == 0 && envValue != "" {
					intV, err := strconv.Atoi(envValue)
					if err != nil {
						log.Error().Err(err).Msgf("unable to convert environment variable to int: %s", envValue)
					}
					intValue = int64(intV)
				}
				field.SetInt(intValue)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue := viper.GetUint64(fieldTag)
			field.SetUint(uintValue)
		case reflect.Float32, reflect.Float64:
			floatValue := viper.GetFloat64(fieldTag)
			field.SetFloat(floatValue)
		case reflect.Struct:
			if fieldTag != "" {
				nestedValue := reflect.New(fieldType).Interface()
				err := viper.UnmarshalKey(fieldTag, nestedValue)
				if err != nil {
					return nil, fmt.Errorf("unable to unmarshal nested struct: %s", err.Error())
				}
				if field.CanSet() {
					field.Set(reflect.ValueOf(nestedValue).Elem())
				}
			} else {
				nestedType := field.Type()
				nestedValue := reflect.New(nestedType).Interface()
				_, err := ReadEnvVarsIntoStruct(nestedValue)
				if err != nil {
					return nil, err
				}
				if field.CanSet() {
					field.Set(reflect.ValueOf(nestedValue).Elem())
				}
			}
		case reflect.Ptr:
			if fieldType.Elem().Kind() == reflect.Struct {
				if field.IsNil() {
					field.Set(reflect.New(fieldType.Elem()))
				}
				nestedValue, err := ReadEnvVarsIntoStruct(field.Interface())
				if err != nil {
					return nil, err
				}
				field.Set(reflect.ValueOf(nestedValue))
			}
		default:
			return nil, fmt.Errorf("unsupported field type %s", fieldType.Kind().String())
		}
	}

	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return nil, err
	}

	return s, nil
}
