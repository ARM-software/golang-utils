/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package slice

import serialisation "github.com/ARM-software/golang-utils/utils/serialisation/slice"

// ToSlice converts a struct to a list of key values.
func ToSlice[T any](s *T) ([]string, error) { return serialisation.ToSlice(s) }

// FromSlice converts a slice of key,values into a struct.
func FromSlice[T any](s []string, o *T) error { return serialisation.FromSlice(s, o) }

// FromArgs converts a list of args into a struct.
func FromArgs[T any](o *T, args ...string) error { return serialisation.FromArgs(o, args...) }
