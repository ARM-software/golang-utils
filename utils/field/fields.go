/*
 * Copyright (C) 2020-2023 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// package field provides utilities to set structure fields. It was inspired by the kubernetes package https://pkg.go.dev/k8s.io/utils/pointer.
package field

import "time"

// ToOptionalInt returns a pointer to an int
func ToOptionalInt(f int) *int {
	return ToOptional(f)
}

// OptionalInt returns the value of an optional field or else
// returns defaultValue.
func OptionalInt(ptr *int, defaultValue int) int {
	return Optional(ptr, defaultValue)
}

// ToOptionalInt32 returns a pointer to an int32.
func ToOptionalInt32(f int32) *int32 {
	return ToOptional(f)
}

// OptionalInt32 returns the value of an optional field or else
// returns defaultValue.
func OptionalInt32(ptr *int32, defaultValue int32) int32 {
	return Optional(ptr, defaultValue)
}

// ToOptionalUint returns a pointer to an uint
func ToOptionalUint(f uint) *uint {
	return ToOptional(f)
}

// OptionalUint returns the value of an optional field or else returns defaultValue.
func OptionalUint(ptr *uint, defaultValue uint) uint {
	return Optional(ptr, defaultValue)
}

// ToOptionalUint32 returns a pointer to an uint32.
func ToOptionalUint32(f uint32) *uint32 {
	return ToOptional(f)
}

// OptionalUint32 returns the value of an optional field or else returns defaultValue.
func OptionalUint32(ptr *uint32, defaultValue uint32) uint32 {
	return Optional(ptr, defaultValue)
}

// ToOptionalInt64 returns a pointer to an int64.
func ToOptionalInt64(f int64) *int64 {
	return ToOptional(f)
}

// OptionalInt64 returns the value of an optional field or else returns defaultValue.
func OptionalInt64(ptr *int64, defaultValue int64) int64 {
	return Optional(ptr, defaultValue)
}

// ToOptionalUint64 returns a pointer to an uint64.
func ToOptionalUint64(f uint64) *uint64 {
	return ToOptional(f)
}

// OptionalUint64 returns the value of an optional field or else returns defaultValue.
func OptionalUint64(ptr *uint64, defaultValue uint64) uint64 {
	return Optional(ptr, defaultValue)
}

// ToOptionalBool returns a pointer to a bool.
func ToOptionalBool(b bool) *bool {
	return ToOptional(b)
}

// OptionalBool returns the value of an optional field or else returns defaultValue.
func OptionalBool(ptr *bool, defaultValue bool) bool {
	return Optional(ptr, defaultValue)
}

// ToOptionalString returns a pointer to a string.
func ToOptionalString(s string) *string {
	return ToOptional(s)
}

// OptionalString returns the value of an optional field or else returns defaultValue.
func OptionalString(ptr *string, defaultValue string) string {
	return Optional(ptr, defaultValue)
}

// ToOptionalAny returns a pointer to a object.
func ToOptionalAny(a any) *any {
	return ToOptional(a)
}

// OptionalAny returns the value of an optional field or else returns defaultValue.
func OptionalAny(ptr *any, defaultValue any) any {
	return Optional(ptr, defaultValue)
}

// ToOptionalFloat32 returns a pointer to a float32.
func ToOptionalFloat32(f float32) *float32 {
	return ToOptional(f)
}

// OptionalFloat32 returns the value of an optional field or else returns defaultValue.
func OptionalFloat32(ptr *float32, defaultValue float32) float32 {
	return Optional(ptr, defaultValue)
}

// ToOptionalFloat64 returns a pointer to a float64.
func ToOptionalFloat64(f float64) *float64 {
	return ToOptional(f)
}

// OptionalFloat64 returns the value of an optional field or else returns defaultValue.
func OptionalFloat64(ptr *float64, defaultValue float64) float64 {
	return Optional(ptr, defaultValue)
}

// ToOptionalDuration returns a pointer to a Duration.
func ToOptionalDuration(f time.Duration) *time.Duration {
	return ToOptional(f)
}

// OptionalDuration returns the value of an optional field or else returns defaultValue.
func OptionalDuration(ptr *time.Duration, defaultValue time.Duration) time.Duration {
	return Optional(ptr, defaultValue)
}

// ToOptionalTime returns a pointer to a Time.
func ToOptionalTime(f time.Time) *time.Time {
	return ToOptional(f)
}

// OptionalTime returns the value of an optional field or else returns defaultValue.
func OptionalTime(ptr *time.Time, defaultValue time.Time) time.Time {
	return Optional(ptr, defaultValue)
}

// ToOptional returns a pointer to the given field value.
func ToOptional[T any](v T) *T {
	return &v
}

// Optional  returns the value of an optional field or else returns defaultValue.
func Optional[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// Equal returns true if both arguments are nil or both arguments
// dereference to the same value.
func Equal[T comparable](a, b *T) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	return EqualValue(a, *b)
}

// EqualValue returns true if optional field dereferences to the value.
func EqualValue[T comparable](field *T, value T) bool {
	if field == nil {
		return false
	}
	return *field == value
}
