/*
 * Copyright (C) 2020-2023 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package field

import (
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/safecast"
)

func TestOptionalField(t *testing.T) {
	tests := []struct {
		fieldType    string
		value        any
		defaultValue any
		setFunction  func(any) any
		getFunction  func(any, any) any
	}{
		{
			fieldType:    "Int",
			value:        safecast.ToInt(time.Now().Second()),
			defaultValue: 76,
			setFunction: func(a any) any {
				return ToOptionalInt(a.(int))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *int
				if a != nil {
					ptr = a.(*int)
				}
				return OptionalInt(ptr, a2.(int))
			},
		},
		{
			fieldType:    "UInt",
			value:        safecast.ToUint(time.Now().Second()),
			defaultValue: safecast.ToUint(76),
			setFunction: func(a any) any {
				return ToOptionalUint(a.(uint))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *uint
				if a != nil {
					ptr = a.(*uint)
				}
				return OptionalUint(ptr, a2.(uint))
			},
		},
		{
			fieldType:    "Int32",
			value:        safecast.ToInt32(time.Now().Second()),
			defaultValue: safecast.ToInt32(97894),
			setFunction: func(a any) any {
				return ToOptionalInt32(a.(int32))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *int32
				if a != nil {
					ptr = a.(*int32)
				}
				return OptionalInt32(ptr, a2.(int32))
			},
		},
		{
			fieldType:    "UInt32",
			value:        safecast.ToUint32(time.Now().Second()),
			defaultValue: safecast.ToUint32(97894),
			setFunction: func(a any) any {
				return ToOptionalUint32(a.(uint32))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *uint32
				if a != nil {
					ptr = a.(*uint32)
				}
				return OptionalUint32(ptr, a2.(uint32))
			},
		},
		{
			fieldType:    "Int64",
			value:        time.Now().Unix(),
			defaultValue: safecast.ToInt64(97894),
			setFunction: func(a any) any {
				return ToOptionalInt64(a.(int64))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *int64
				if a != nil {
					ptr = a.(*int64)
				}
				return OptionalInt64(ptr, a2.(int64))
			},
		},
		{
			fieldType:    "UInt64",
			value:        safecast.ToUint64(time.Now().Unix()),
			defaultValue: safecast.ToUint64(97894),
			setFunction: func(a any) any {
				return ToOptionalUint64(a.(uint64))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *uint64
				if a != nil {
					ptr = a.(*uint64)
				}
				return OptionalUint64(ptr, a2.(uint64))
			},
		},
		{
			fieldType:    "Float32",
			value:        float32(time.Now().Second()),
			defaultValue: float32(97894.1545),
			setFunction: func(a any) any {
				return ToOptionalFloat32(a.(float32))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *float32
				if a != nil {
					ptr = a.(*float32)
				}
				return OptionalFloat32(ptr, a2.(float32))
			},
		},
		{
			fieldType:    "Float64",
			value:        float64(time.Now().Second()),
			defaultValue: float64(97894.1545),
			setFunction: func(a any) any {
				return ToOptionalFloat64(a.(float64))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *float64
				if a != nil {
					ptr = a.(*float64)
				}
				return OptionalFloat64(ptr, a2.(float64))
			},
		},
		{
			fieldType:    "Bool",
			value:        false,
			defaultValue: true,
			setFunction: func(a any) any {
				return ToOptionalBool(a.(bool))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *bool
				if a != nil {
					ptr = a.(*bool)
				}
				return OptionalBool(ptr, a2.(bool))
			},
		},
		{
			fieldType:    "String",
			value:        faker.Sentence(),
			defaultValue: faker.Name(),
			setFunction: func(a any) any {
				return ToOptionalString(a.(string))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *string
				if a != nil {
					ptr = a.(*string)
				}
				return OptionalString(ptr, a2.(string))
			},
		},
		{
			fieldType:    "Duration",
			value:        time.Millisecond,
			defaultValue: time.Second,
			setFunction: func(a any) any {
				return ToOptionalDuration(a.(time.Duration))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *time.Duration
				if a != nil {
					ptr = a.(*time.Duration)
				}
				return OptionalDuration(ptr, a2.(time.Duration))
			},
		},
		{
			fieldType:    "time",
			value:        time.Now(),
			defaultValue: time.Now().Add(-time.Hour),
			setFunction: func(a any) any {
				return ToOptionalTime(a.(time.Time))
			},
			getFunction: func(a any, a2 any) any {
				var ptr *time.Time
				if a != nil {
					ptr = a.(*time.Time)
				}
				return OptionalTime(ptr, a2.(time.Time))
			},
		},
		{
			fieldType:    "Any",
			value:        faker.Sentence(),
			defaultValue: time.Now(),
			setFunction: func(a any) any {
				return ToOptionalAny(a)
			},
			getFunction: func(a any, a2 any) any {
				var ptr *any
				if a != nil {
					ptr = a.(*any)
				}
				return OptionalAny(ptr, a2)
			},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.fieldType, func(t *testing.T) {
			to := test.setFunction(test.value)
			assert.NotNil(t, to)
			assert.Equal(t, test.defaultValue, test.getFunction(nil, test.defaultValue))
			assert.Equal(t, test.value, test.getFunction(to, test.defaultValue))
		})
	}
}

func TestEqual(t *testing.T) {
	type T int
	assert.True(t, Equal[T](nil, nil))
	assert.False(t, Equal(ToOptionalInt(1), nil))
	assert.False(t, Equal(nil, ToOptionalInt(1)))
	assert.True(t, Equal(ToOptionalInt(1), ToOptionalInt(1)))
	str := faker.Sentence()
	assert.True(t, Equal(ToOptionalString(str), ToOptionalString(str)))
	assert.False(t, Equal(ToOptionalString(str), ToOptionalString("")))
}

func TestEqualValue(t *testing.T) {
	type T any
	assert.False(t, EqualValue[T](nil, nil))
	assert.False(t, EqualValue(nil, 1))
	assert.True(t, EqualValue(ToOptionalInt(1), 1))
	str := faker.Sentence()
	assert.True(t, EqualValue(ToOptionalString(str), str))
	assert.False(t, EqualValue(ToOptionalString(str), ""))
}
