/*
 * Copyright (C) 2020-2023 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// package field provides utilities to set structure fields. It was inspired by the kubernetes package https://pkg.go.dev/k8s.io/utils/pointer.
package field

import (
	"time"

	"github.com/ARM-software/golang-utils/utils/value"
)

// ToOptionalInt returns a pointer to an int
func ToOptionalInt(f int) *int {
	return ToOptional[int](f)
}

// ToOptionalIntOrNilIfEmpty returns a pointer to an int unless it is empty and in that case returns nil.
func ToOptionalIntOrNilIfEmpty(f int) *int {
	return ToOptionalOrNilIfEmpty[int](f)
}

// OptionalInt returns the value of an optional field or else
// returns defaultValue.
func OptionalInt(ptr *int, defaultValue int) int {
	return Optional[int](ptr, defaultValue)
}

// ToOptionalInt32 returns a pointer to an int32.
func ToOptionalInt32(f int32) *int32 {
	return ToOptional[int32](f)
}

// ToOptionalInt32OrNilIfEmpty returns a pointer to an int32 unless it is empty and in that case returns nil.
func ToOptionalInt32OrNilIfEmpty(f int32) *int32 {
	return ToOptionalOrNilIfEmpty[int32](f)
}

// OptionalInt32 returns the value of an optional field or else
// returns defaultValue.
func OptionalInt32(ptr *int32, defaultValue int32) int32 {
	return Optional[int32](ptr, defaultValue)
}

// ToOptionalUint returns a pointer to an uint
func ToOptionalUint(f uint) *uint {
	return ToOptional[uint](f)
}

// ToOptionalUintOrNilIfEmpty returns a pointer to a Uint unless it is empty and in that case returns nil.
func ToOptionalUintOrNilIfEmpty(f uint) *uint {
	return ToOptionalOrNilIfEmpty[uint](f)
}

// OptionalUint returns the value of an optional field or else returns defaultValue.
func OptionalUint(ptr *uint, defaultValue uint) uint {
	return Optional[uint](ptr, defaultValue)
}

// ToOptionalUint32 returns a pointer to an uint32.
func ToOptionalUint32(f uint32) *uint32 {
	return ToOptional[uint32](f)
}

// ToOptionalUint32OrNilIfEmpty returns a pointer to an Uint32 unless it is empty and in that case returns nil.
func ToOptionalUint32OrNilIfEmpty(f uint32) *uint32 {
	return ToOptionalOrNilIfEmpty[uint32](f)
}

// OptionalUint32 returns the value of an optional field or else returns defaultValue.
func OptionalUint32(ptr *uint32, defaultValue uint32) uint32 {
	return Optional[uint32](ptr, defaultValue)
}

// ToOptionalInt64 returns a pointer to an int64.
func ToOptionalInt64(f int64) *int64 {
	return ToOptional[int64](f)
}

// ToOptionalInt64OrNilIfEmpty returns a pointer to an int64 unless it is empty and in that case returns nil.
func ToOptionalInt64OrNilIfEmpty(f int64) *int64 {
	return ToOptionalOrNilIfEmpty[int64](f)
}

// OptionalInt64 returns the value of an optional field or else returns defaultValue.
func OptionalInt64(ptr *int64, defaultValue int64) int64 {
	return Optional[int64](ptr, defaultValue)
}

// ToOptionalUint64 returns a pointer to an uint64.
func ToOptionalUint64(f uint64) *uint64 {
	return ToOptional[uint64](f)
}

// ToOptionalUint64OrNilIfEmpty returns a pointer to an Uint64 unless it is empty and in that case returns nil.
func ToOptionalUint64OrNilIfEmpty(f uint64) *uint64 {
	return ToOptionalOrNilIfEmpty[uint64](f)
}

// OptionalUint64 returns the value of an optional field or else returns defaultValue.
func OptionalUint64(ptr *uint64, defaultValue uint64) uint64 {
	return Optional[uint64](ptr, defaultValue)
}

// ToOptionalBool returns a pointer to a bool.
func ToOptionalBool(b bool) *bool {
	return ToOptional[bool](b)
}

// ToOptionalBoolOrNilIfEmpty returns a pointer to a boolean unless it is empty and in that case returns nil.
func ToOptionalBoolOrNilIfEmpty(f bool) *bool {
	return ToOptionalOrNilIfEmpty[bool](f)
}

// OptionalBool returns the value of an optional field or else returns defaultValue.
func OptionalBool(ptr *bool, defaultValue bool) bool {
	return Optional[bool](ptr, defaultValue)
}

// ToOptionalString returns a pointer to a string.
func ToOptionalString(s string) *string {
	return ToOptional[string](s)
}

// ToOptionalStringOrNilIfEmpty returns a pointer to a string unless it is empty and in that case returns nil.
func ToOptionalStringOrNilIfEmpty(f string) *string {
	return ToOptionalOrNilIfEmpty[string](f)
}

// OptionalString returns the value of an optional field or else returns defaultValue.
func OptionalString(ptr *string, defaultValue string) string {
	return Optional[string](ptr, defaultValue)
}

// ToOptionalAny returns a pointer to an object.
func ToOptionalAny(a any) *any {
	return ToOptional[any](a)
}

// ToOptionalAnyOrNilIfEmpty returns a pointer to an object unless it is empty and in that case returns nil.
func ToOptionalAnyOrNilIfEmpty(f any) *any {
	return ToOptionalOrNilIfEmpty[any](f)
}

// OptionalAny returns the value of an optional field or else returns defaultValue.
func OptionalAny(ptr *any, defaultValue any) any {
	return Optional[any](ptr, defaultValue)
}

// ToOptionalFloat32 returns a pointer to a float32.
func ToOptionalFloat32(f float32) *float32 {
	return ToOptional[float32](f)
}

// ToOptionalFloat32OrNilIfEmpty returns a pointer to a float32 unless it is empty and in that case returns nil.
func ToOptionalFloat32OrNilIfEmpty(f float32) *float32 {
	return ToOptionalOrNilIfEmpty[float32](f)
}

// OptionalFloat32 returns the value of an optional field or else returns defaultValue.
func OptionalFloat32(ptr *float32, defaultValue float32) float32 {
	return Optional(ptr, defaultValue)
}

// ToOptionalFloat64 returns a pointer to a float64.
func ToOptionalFloat64(f float64) *float64 {
	return ToOptional[float64](f)
}

// ToOptionalFloat64OrNilIfEmpty returns a pointer to a float64 unless it is empty and in that case returns nil.
func ToOptionalFloat64OrNilIfEmpty(f float64) *float64 {
	return ToOptionalOrNilIfEmpty[float64](f)
}

// OptionalFloat64 returns the value of an optional field or else returns defaultValue.
func OptionalFloat64(ptr *float64, defaultValue float64) float64 {
	return Optional[float64](ptr, defaultValue)
}

// ToOptionalDuration returns a pointer to a Duration.
func ToOptionalDuration(f time.Duration) *time.Duration {
	return ToOptional[time.Duration](f)
}

// ToOptionalDurationOrNilIfEmpty returns a pointer to a duration unless it is empty and in that case returns nil.
func ToOptionalDurationOrNilIfEmpty(f time.Duration) *time.Duration {
	return ToOptionalOrNilIfEmpty[time.Duration](f)
}

// OptionalDuration returns the value of an optional field or else returns defaultValue.
func OptionalDuration(ptr *time.Duration, defaultValue time.Duration) time.Duration {
	return Optional[time.Duration](ptr, defaultValue)
}

// ToOptionalTime returns a pointer to a Time.
func ToOptionalTime(f time.Time) *time.Time {
	return ToOptional[time.Time](f)
}

// ToOptionalTimeOrNilIfEmpty returns a pointer to a time unless it is empty and in that case returns nil.
func ToOptionalTimeOrNilIfEmpty(f time.Time) *time.Time {
	return ToOptionalOrNilIfEmpty[time.Time](f)
}

// OptionalTime returns the value of an optional field or else returns defaultValue.
func OptionalTime(ptr *time.Time, defaultValue time.Time) time.Time {
	return Optional[time.Time](ptr, defaultValue)
}

// ToOptional returns a pointer to the given field value.
func ToOptional[T any](v T) *T {
	return &v
}

// ToOptionalOrNilIfEmpty returns a pointer to the given field value unless it is empty and in that case returns nil.
func ToOptionalOrNilIfEmpty[T any](v T) *T {
	if value.IsEmpty(v) {
		return nil
	}
	return ToOptional[T](v)
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
