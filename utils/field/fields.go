/*
 * Copyright (C) 2020-2023 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// package field provides utilities to set structure fields. It was inspired by the kubernetes package https://pkg.go.dev/k8s.io/utils/pointer.
package field

import "time"

// ToOptionalInt returns a pointer to an int
func ToOptionalInt(i int) *int {
	return &i
}

// OptionalInt returns the value of an optional field or else
// returns defaultValue.
func OptionalInt(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalInt32 returns a pointer to an int32.
func ToOptionalInt32(i int32) *int32 {
	return &i
}

// OptionalInt32 returns the value of an optional field or else
// returns defaultValue.
func OptionalInt32(ptr *int32, defaultValue int32) int32 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalUint returns a pointer to an uint
func ToOptionalUint(i uint) *uint {
	return &i
}

// OptionalUint returns the value of an optional field or else returns defaultValue.
func OptionalUint(ptr *uint, defaultValue uint) uint {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalUint32 returns a pointer to an uint32.
func ToOptionalUint32(i uint32) *uint32 {
	return &i
}

// OptionalUint32 returns the value of an optional field or else returns defaultValue.
func OptionalUint32(ptr *uint32, defaultValue uint32) uint32 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalInt64 returns a pointer to an int64.
func ToOptionalInt64(i int64) *int64 {
	return &i
}

// OptionalInt64 returns the value of an optional field or else returns defaultValue.
func OptionalInt64(ptr *int64, defaultValue int64) int64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalUint64 returns a pointer to an uint64.
func ToOptionalUint64(i uint64) *uint64 {
	return &i
}

// OptionalUint64 returns the value of an optional field or else returns defaultValue.
func OptionalUint64(ptr *uint64, defaultValue uint64) uint64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalBool returns a pointer to a bool.
func ToOptionalBool(b bool) *bool {
	return &b
}

// OptionalBool returns the value of an optional field or else returns defaultValue.
func OptionalBool(ptr *bool, defaultValue bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalString returns a pointer to a string.
func ToOptionalString(s string) *string {
	return &s
}

// OptionalString returns the value of an optional field or else returns defaultValue.
func OptionalString(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalAny returns a pointer to a object.
func ToOptionalAny(a any) *any {
	return &a
}

// OptionalAny returns the value of an optional field or else returns defaultValue.
func OptionalAny(ptr *any, defaultValue any) any {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalFloat32 returns a pointer to a float32.
func ToOptionalFloat32(i float32) *float32 {
	return &i
}

// OptionalFloat32 returns the value of an optional field or else returns defaultValue.
func OptionalFloat32(ptr *float32, defaultValue float32) float32 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalFloat64 returns a pointer to a float64.
func ToOptionalFloat64(i float64) *float64 {
	return &i
}

// OptionalFloat64 returns the value of an optional field or else returns defaultValue.
func OptionalFloat64(ptr *float64, defaultValue float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalDuration returns a pointer to a Duration.
func ToOptionalDuration(i time.Duration) *time.Duration {
	return &i
}

// OptionalDuration returns the value of an optional field or else returns defaultValue.
func OptionalDuration(ptr *time.Duration, defaultValue time.Duration) time.Duration {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// ToOptionalTime returns a pointer to a Time.
func ToOptionalTime(i time.Time) *time.Time {
	return &i
}

// OptionalTime returns the value of an optional field or else returns defaultValue.
func OptionalTime(ptr *time.Time, defaultValue time.Time) time.Time {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
