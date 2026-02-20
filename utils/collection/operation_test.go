/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/safecast"
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

func TestFindInSequence(t *testing.T) {

	index, found := FindInSequence[string](slices.Values([]string{}), func(e string) bool {
		return e == "D"
	})
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSequence[string](slices.Values([]string{}), func(_ string) bool {
		return true
	})
	assert.False(t, found)
	assert.Equal(t, -1, index)

	index, found = FindInSequence[string](slices.Values([]string{"A", "b", "c"}), func(e string) bool {
		return e == "D"
	})
	assert.False(t, found)
	assert.Equal(t, -1, index)
	index, found = FindInSequenceRef[string](slices.Values([]string{"A", "b", "c"}), func(e *string) bool {
		return field.OptionalString(e, "") == "D"
	})
	assert.False(t, found)
	assert.Equal(t, -1, index)

	index, found = FindInSequence[string](slices.Values([]string{"A", "B", "b", "c"}), func(e string) bool {
		return e == "b"
	})
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

	entry := []int{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	intValues := UniqueEntries(entry)
	assert.Len(t, intValues, 4)

	intValues = Unique(slices.Values(entry))
	assert.Len(t, intValues, 4)
}

func TestSetOperations(t *testing.T) {
	a := []string{"a", "b", "b", "c"}
	b := []string{"b", "c", "d", "d"}

	assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, Union[string](a, b))
	assert.ElementsMatch(t, UniqueEntries[string](a), Union[string](a, nil))
	assert.ElementsMatch(t, UniqueEntries[string](b), Union[string](nil, b))
	assert.Empty(t, Union[string](nil, nil))
	assert.ElementsMatch(t, []string{"b", "c"}, Intersection[string](a, b))
	assert.Empty(t, Intersection[string](nil, b))
	assert.Empty(t, Intersection[string](nil, nil))
	assert.Empty(t, Intersection[string](a, nil))
	assert.ElementsMatch(t, []string{"a"}, Difference[string](a, b))
	assert.Empty(t, Difference[string](nil, b))
	assert.Empty(t, Difference[string](nil, nil))
	assert.ElementsMatch(t, UniqueEntries[string](a), Difference[string](a, nil))
	assert.ElementsMatch(t, []string{"a", "d"}, SymmetricDifference[string](a, b))
	assert.ElementsMatch(t, UniqueEntries[string](a), SymmetricDifference[string](a, nil))
	assert.Empty(t, SymmetricDifference[string](nil, nil))
	assert.ElementsMatch(t, UniqueEntries[string](b), SymmetricDifference[string](nil, b))

	entry := []int{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	assert.ElementsMatch(t, UniqueEntries[int](entry), Union[int](entry, entry))
	assert.ElementsMatch(t, UniqueEntries[int](entry), Intersection[int](entry, entry))
	assert.Empty(t, Difference[int](entry, entry))
	assert.Empty(t, SymmetricDifference[int](entry, entry))
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
	assert.True(t, AnyRefFunc([]bool{true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true}, func(b *bool) bool {
		return f(field.OptionalBool(b, false))
	}))
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
	assert.ElementsMatch(t, []int{2, 4}, FilterRef(nums, func(n *int) bool {
		return field.OptionalInt(n, 0)%2 == 0
	}))
	assert.ElementsMatch(t, []int{1, 3, 5}, Reject(nums, func(n int) bool {
		return n%2 == 0
	}))
	assert.ElementsMatch(t, []int{1, 3, 5}, RejectRef(nums, func(n *int) bool {
		return field.OptionalInt(n, 0)%2 == 0
	}))
	assert.ElementsMatch(t, []int{1, 3, 5}, slices.Collect[int](RejectSequence[int](slices.Values(nums), func(n int) bool {
		return n%2 == 0
	})))
	assert.ElementsMatch(t, []int{4, 5}, Filter(nums, func(n int) bool {
		return n > 3
	}))
	assert.ElementsMatch(t, []int{4, 5}, FilterRef(nums, func(n *int) bool {
		return *n > 3
	}))
	assert.ElementsMatch(t, []int{4, 5}, slices.Collect(FilterRefSequence(slices.Values(nums), func(n *int) bool {
		return *n > 3
	})))
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

func TestMatch(t *testing.T) {
	match1 := func(i int) bool { return i == 1 }
	match2 := func(i int) bool { return i == 2 }
	match3 := func(i int) bool { return i == 3 }
	assert.True(t, Match(1, match1, match2, match3))
	assert.True(t, Match(2, match1, match2, match3))
	assert.True(t, Match(3, match1, match2, match3))
	assert.False(t, Match(4, match1, match2, match3))
	assert.False(t, Match(0, match1, match2, match3))
	assert.False(t, Match(2, match1, match3))
	assert.True(t, MatchAll(1, match1))
	assert.False(t, MatchAll(1, match1, match2))
}

func TestMap(t *testing.T) {
	mapped := Map([]int{1, 2}, func(i int) string {
		return fmt.Sprintf("Hello world %v", i)
	})
	assert.ElementsMatch(t, []string{"Hello world 1", "Hello world 2"}, mapped)
	num := []int{1, 2, 3, 4}
	numStr := []string{"1", "2", "3", "4"}
	mapped = Map(num, func(x int) string {
		return strconv.FormatInt(safecast.ToInt64(x), 10)
	})
	assert.ElementsMatch(t, numStr, mapped)
	mapped = MapRef(num, func(x *int) *string {
		return field.ToOptionalOrNilIfEmpty(strconv.FormatInt(safecast.ToInt64(field.OptionalInt(x, 0)), 10))
	})
	assert.ElementsMatch(t, numStr, mapped)
	m, err := MapWithError[string, int](numStr, strconv.Atoi)
	require.NoError(t, err)
	assert.ElementsMatch(t, num, m)
	_, err = MapWithError[string, int](append(numStr, faker.Word(), "5"), strconv.Atoi)
	require.Error(t, err)
	_, err = MapRefWithError[string, int](append(numStr, faker.Word(), "5"), func(s *string) (*int, error) {
		if s == nil {
			return nil, commonerrors.ErrUndefined
		}
		n, err := strconv.Atoi(*s)
		return &n, err
	})
	require.Error(t, err)

	mappedInt := Map[int, int](num, IdentityMapFunc[int]())
	assert.ElementsMatch(t, num, mappedInt)
}

func TestReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sumOfNums := Reduce(nums, 0, func(acc, n int) int {
		return acc + n
	})
	assert.Equal(t, sumOfNums, 15)
}

func TestForEach(t *testing.T) {
	list := Range(9, 1000, field.ToOptionalInt(13))
	t.Run("each", func(t *testing.T) {
		var visited []int
		ForEachValues(func(i int) {
			visited = append(visited, i)
		}, list...)
		assert.ElementsMatch(t, visited, list)
	})
	t.Run("foreachref", func(t *testing.T) {
		var visited []int
		ForEachRef(list, func(i *int) {
			visited = append(visited, field.OptionalInt(i, 0))
		})
		assert.ElementsMatch(t, visited, list)
	})
	t.Run("eachref", func(t *testing.T) {
		var visited []int
		require.NoError(t, EachRef(slices.Values(list), func(i *int) error {
			visited = append(visited, field.OptionalInt(i, 0))
			return nil
		}))
		assert.ElementsMatch(t, visited, list)
	})
}

func TestForEachSequence(t *testing.T) {
	var visited []int
	list := Range(9, 1000, field.ToOptionalInt(13))
	errortest.AssertError(t, Each(slices.Values(list), func(i int) error {
		if i > 150 {
			return commonerrors.ErrUnsupported
		}
		visited = append(visited, i)
		return nil
	}), commonerrors.ErrUnsupported)
	assert.ElementsMatch(t, visited, []int{9, 22, 35, 48, 61, 74, 87, 100, 113, 126, 139})
	visited = []int{}
	assert.NoError(t, Each(slices.Values(list), func(i int) error {
		if i > 150 {
			return commonerrors.ErrEOF
		}
		visited = append(visited, i)
		return nil
	}))
	assert.ElementsMatch(t, visited, []int{9, 22, 35, 48, 61, 74, 87, 100, 113, 126, 139})
}

func TestForAll(t *testing.T) {
	var visited []int
	list := Range(9, 1000, field.ToOptionalInt(13))
	errortest.AssertError(t, ForAll(list, func(i int) error {
		visited = append(visited, i)
		if i > 150 {
			return commonerrors.ErrUnsupported
		}
		return nil
	}), commonerrors.ErrUnsupported)
	assert.ElementsMatch(t, visited, []int{9, 22, 35, 48, 61, 74, 87, 100, 113, 126, 139, 152, 165, 178, 191, 204, 217, 230, 243, 256, 269, 282, 295, 308, 321, 334, 347, 360, 373, 386, 399, 412, 425, 438, 451, 464, 477, 490, 503, 516, 529, 542, 555, 568, 581, 594, 607, 620, 633, 646, 659, 672, 685, 698, 711, 724, 737, 750, 763, 776, 789, 802, 815, 828, 841, 854, 867, 880, 893, 906, 919, 932, 945, 958, 971, 984, 997})
	visited = []int{}
	assert.NoError(t, ForAll(list, func(i int) error {
		visited = append(visited, i)
		if i > 150 {
			return commonerrors.ErrEOF
		}
		return nil
	}))
	assert.ElementsMatch(t, visited, []int{9, 22, 35, 48, 61, 74, 87, 100, 113, 126, 139, 152})
	visited = []int{}
	assert.NoError(t, ForAllRef(list, func(i *int) error {
		n := field.OptionalInt(i, 0)
		visited = append(visited, n)
		if n > 150 {
			return commonerrors.ErrEOF
		}
		return nil
	}))
	assert.ElementsMatch(t, visited, []int{9, 22, 35, 48, 61, 74, 87, 100, 113, 126, 139, 152})
	visited = []int{}
	assert.NoError(t, ForAllSequenceRef(slices.Values(list), func(i *int) error {
		n := field.OptionalInt(i, 0)
		visited = append(visited, n)
		if n > 150 {
			return commonerrors.ErrEOF
		}
		return nil
	}))
	assert.ElementsMatch(t, visited, []int{9, 22, 35, 48, 61, 74, 87, 100, 113, 126, 139, 152})
	assert.NoError(t, ForAll(list, func(i int) error {
		return nil
	}))
	assert.NoError(t, ForAllRef(list, func(i *int) error {
		return nil
	}))
}
