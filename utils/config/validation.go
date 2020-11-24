package config

import "reflect"

type Validator interface {
	Validate() error
}

// Use reflection to find embedded structs and validate them
func ValidateEmbedded(cfg Validator) error {
	r := reflect.ValueOf(cfg).Elem()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		if f.Kind() == reflect.Struct {
			validator, ok := f.Addr().Interface().(Validator)
			if !ok {
				continue
			}
			err := validator.Validate()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
