/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package maps

import (
	"github.com/go-viper/mapstructure/v2"

	serialisation "github.com/ARM-software/golang-utils/utils/serialisation/maps"
)

// ToMapFromPointer converts a pointer to a flat map.
func ToMapFromPointer[T any](o T) (map[string]string, error) {
	return serialisation.ToMapFromPointer(o)
}

// ToMap converts a struct to a flat map.
func ToMap[T any](o *T) (map[string]string, error) { return serialisation.ToMap(o) }

// FromMapToPointer converts a flat map into a pointer.
func FromMapToPointer[T any](m map[string]string, o T) error {
	return serialisation.FromMapToPointer(m, o)
}

// FromMap converts a flat map into a struct.
func FromMap[T any](m map[string]string, o *T) error { return serialisation.FromMap(m, o) }

// CustomTypeHookFunc returns the custom type decode hook.
func CustomTypeHookFunc() mapstructure.DecodeHookFunc { return serialisation.CustomTypeHookFunc() }
