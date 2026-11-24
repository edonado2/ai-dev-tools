package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

func validate(cfg Config) error {
	return validateStruct(cfg)
}

func validateStruct(cfg any) error {
	v := reflect.ValueOf(cfg)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get validate tag
		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			// Recursively validate nested structs
			if field.Kind() == reflect.Struct {
				if err := validateStruct(field.Interface()); err != nil {
					return err
				}
			}
			continue
		}

		// Parse validation rules
		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if err := validateField(field, fieldType.Name, rule); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateField(field reflect.Value, fieldName, rule string) error {
	switch {
	case strings.HasPrefix(rule, "oneof="):
		allowed := strings.Split(strings.TrimPrefix(rule, "oneof="), " ")
		val := field.String()
		found := false
		for _, a := range allowed {
			if val == a {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s: must be one of %v, got %q", fieldName, allowed, val)
		}
	case strings.HasPrefix(rule, "min="):
		minVal := parseInt(strings.TrimPrefix(rule, "min="))
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() < int64(minVal) {
				return fmt.Errorf("%s: must be >= %d, got %d", fieldName, minVal, field.Int())
			}
		case reflect.String:
			if len(field.String()) < minVal {
				return fmt.Errorf("%s: length must be >= %d, got %d", fieldName, minVal, len(field.String()))
			}
		}
	case strings.HasPrefix(rule, "max="):
		maxVal := parseInt(strings.TrimPrefix(rule, "max="))
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() > int64(maxVal) {
				return fmt.Errorf("%s: must be <= %d, got %d", fieldName, maxVal, field.Int())
			}
		case reflect.String:
			if len(field.String()) > maxVal {
				return fmt.Errorf("%s: length must be <= %d, got %d", fieldName, maxVal, len(field.String()))
			}
		}
	}
	return nil
}

func parseInt(s string) int {
	var result int
	_, _ = fmt.Sscanf(s, "%d", &result)
	return result
}

// DecodeHook for mapstructure to handle string to int conversion
func DecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)
}