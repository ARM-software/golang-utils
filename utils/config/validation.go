/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"reflect"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	fieldUtils "github.com/ARM-software/golang-utils/utils/field"
)

var specialMapstructureTags = []string{"squash", "remain", "omitempty", "omitzero"} // See https://pkg.go.dev/github.com/go-viper/mapstructure/v2#section-readme

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
	mapStructure := fieldUtils.ToOptionalStringOrNilIfEmpty(processMapStructureString(mapStructureStr))
	if !hasTag {
		mapStructure = nil
	}
	err = WrapFieldValidationError(field.Name, mapStructure, nil, err)
	return err
}

// mapstructure has some special tags which need to be accounted for.
func processMapStructureString(str string) string {
	processedStr := strings.TrimSpace(str)
	if processedStr == "-" {
		return ""
	}

	elements := strings.Split(processedStr, ",")
	if len(elements) == 1 {
		return processedStr
	}
	elements = collection.GenericRemove(func(str1, str2 string) bool {
		return strings.EqualFold(strings.TrimSpace(str1), strings.TrimSpace(str2))
	}, elements, specialMapstructureTags...)
	return strings.TrimSpace(strings.Join(elements, ","))
}
