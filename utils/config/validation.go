/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"reflect"
)

// ValidateEmbedded uses reflection to find embedded structs and validate them
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
			field := r.Type().Field(i)

			err = wrapFieldValidationError(field, err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func wrapFieldValidationError(field reflect.StructField, err error) error {
	mapStructureStr, hasTag := field.Tag.Lookup("mapstructure")
	mapStructure := &mapStructureStr
	if !hasTag {
		mapStructure = nil
	}
	err = WrapFieldValidationError(field.Name, mapStructure, nil, err)
	return err
}
