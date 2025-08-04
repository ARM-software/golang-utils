/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	index, found := Find(nil, "D")
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = Find(&[]string{"A", "b", "c"}, "D")
	assert.False(t, found)
	assert.Equal(t, -1, index)

	index, found = Find(&[]string{"A", "B", "b", "c"}, "b")
	assert.True(t, found)
	assert.Equal(t, 2, index)
}

func TestFindInSlice(t *testing.T) {
	index, found := FindInSlice(true, nil, "D")
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSlice(true, []string{"A", "b", "c"})
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSlice(true, []string{"A", "b", "c"}, "D")
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSlice(true, []string{"A", "b", "c"}, "D", "e", "f", "g", "H")
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSlice(false, []string{"A", "b", "c"}, "D")
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSlice(false, []string{"A", "b", "c"}, "D", "e", "f", "g", "H")
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSlice(true, []string{"A", "B", "b", "c"}, "b")
	assert.True(t, found)
	assert.Equal(t, 2, index)
	index, found = FindInSlice(false, []string{"A", "B", "b", "c"}, "b")
	assert.True(t, found)
	assert.Equal(t, 1, index)
	index, found = FindInSlice(true, []string{"A", "B", "b", "c"}, "b", "D", "e", "f", "g", "H")
	assert.True(t, found)
	assert.Equal(t, 2, index)
	index, found = FindInSlice(false, []string{"A", "B", "b", "c"}, "b", "D", "e", "f", "g", "H")
	assert.True(t, found)
	assert.Equal(t, 1, index)
}
func TestUniqueEntries(t *testing.T) {
	assert.Len(t, UniqueEntries([]string{faker.Username(), faker.Name(), faker.Sentence(), faker.Name()}), 4)
	values := UniqueEntries([]string{"test1", "test12", "test1", "test1", "test12", "test12"})
	assert.Len(t, values, 2)
	_, found := FindInSlice(true, values, "test1")
	assert.True(t, found)
	_, found = FindInSlice(true, values, "test12")
	assert.True(t, found)

	intValues := UniqueEntries([]int{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4})
	assert.Len(t, intValues, 4)
}

func TestAnyFunc(t *testing.T) {
	f := func(v bool) bool {
		return v
	}
	assert.False(t, AnyFunc([]bool{}, f))
	assert.True(t, AnyFunc([]bool{false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false}, f))
	assert.False(t, AnyFunc([]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}, f))
	assert.True(t, AnyFunc([]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}, f))
	assert.True(t, AnyFunc([]bool{true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true}, f))
}

func TestAllFunc(t *testing.T) {
	f := func(v bool) bool {
		return v
	}
	assert.False(t, AllFunc([]bool{}, f))
	assert.False(t, AllFunc([]bool{false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false}, f))
	assert.False(t, AllFunc([]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}, f))
	assert.True(t, AllFunc([]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}, f))
	assert.False(t, AllFunc([]bool{true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true}, f))
}

func TestAnyEmpty(t *testing.T) {
	assert.True(t, AnyEmpty(false, []string{faker.Username(), faker.Name(), faker.Sentence(), ""}))
	assert.False(t, AnyEmpty(false, []string{faker.Username(), "         ", faker.Name(), faker.Sentence()}))
	assert.True(t, AnyEmpty(true, []string{faker.Username(), "         ", faker.Name(), faker.Sentence()}))
	assert.True(t, AnyEmpty(false, []string{"", faker.Name(), faker.Sentence()}))
	assert.True(t, AnyEmpty(false, []string{faker.Username(), "", faker.Name(), faker.Sentence()}))
	assert.True(t, AnyEmpty(false, []string{faker.Username(), "", faker.Name(), "", faker.Sentence()}))
	assert.False(t, AnyEmpty(false, []string{faker.Username(), faker.Name(), faker.Sentence()}))
}

func TestAllNotEmpty(t *testing.T) {
	assert.False(t, AllNotEmpty(false, []string{faker.Username(), faker.Name(), faker.Sentence(), ""}))
	assert.False(t, AllNotEmpty(false, []string{"", faker.Name(), faker.Sentence()}))
	assert.False(t, AllNotEmpty(false, []string{faker.Username(), "", faker.Name(), faker.Sentence()}))
	assert.True(t, AllNotEmpty(false, []string{faker.Username(), "     ", faker.Name(), faker.Sentence()}))
	assert.False(t, AllNotEmpty(true, []string{faker.Username(), "      ", faker.Name(), faker.Sentence()}))
	assert.False(t, AllNotEmpty(false, []string{faker.Username(), "", faker.Name(), "", faker.Sentence()}))
	assert.True(t, AllNotEmpty(false, []string{faker.Username(), faker.Name(), faker.Sentence()}))
}

func TestFilterReject(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	assert.ElementsMatch(t, []int{2, 4}, Filter(nums, func(n int) bool {
		return n%2 == 0
	}))
	assert.ElementsMatch(t, []int{1, 3, 5}, Reject(nums, func(n int) bool {
		return n%2 == 0
	}))
	assert.ElementsMatch(t, []int{4, 5}, Filter(nums, func(n int) bool {
		return n > 3
	}))
	assert.ElementsMatch(t, []int{1, 2, 3}, Reject(nums, func(n int) bool {
		return n > 3
	}))
	assert.ElementsMatch(t, []string{"foo", "bar"}, Filter([]string{"", "foo", "", "bar", ""}, func(x string) bool {
		return len(x) > 0
	}))
	assert.ElementsMatch(t, []string{"", "", ""}, Reject([]string{"", "foo", "", "bar", ""}, func(x string) bool {
		return len(x) > 0
	}))
}

func TestMap(t *testing.T) {
	mapped := Map([]int{1, 2}, func(i int) string {
		return fmt.Sprintf("Hello world %v", i)
	})
	assert.ElementsMatch(t, []string{"Hello world 1", "Hello world 2"}, mapped)
	mapped = Map([]int64{1, 2, 3, 4}, func(x int64) string {
		return strconv.FormatInt(x, 10)
	})
	assert.ElementsMatch(t, []string{"1", "2", "3", "4"}, mapped)
}

func TestReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sumOfNums := Reduce(nums, 0, func(acc, n int) int {
		return acc + n
	})
	assert.Equal(t, sumOfNums, 15)
}
