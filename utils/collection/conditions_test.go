package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestAny(t *testing.T) {
	assert.False(t, Any([]bool{}))
	assert.True(t, Any([]bool{false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false}))
	assert.False(t, Any([]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}))
	assert.True(t, Any([]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}))
	assert.True(t, Any([]bool{true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true}))
}

func TestAnyTrue(t *testing.T) {
	assert.False(t, AnyTrue())
	assert.True(t, AnyTrue(false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false))
	assert.False(t, AnyTrue(false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false))
	assert.True(t, AnyTrue(true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true))
	assert.True(t, AnyTrue(true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true))
}

func TestAnyFalse(t *testing.T) {
	assert.True(t, AnyFalse())
	assert.True(t, AnyFalse(false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false))
	assert.True(t, AnyFalse(false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false))
	assert.False(t, AnyFalse(true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true))
	assert.True(t, AnyFalse(true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true))
}

func TestAll(t *testing.T) {
	assert.False(t, All([]bool{}))
	assert.False(t, All([]bool{false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false}))
	assert.False(t, All([]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}))
	assert.True(t, All([]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}))
	assert.False(t, All([]bool{true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true}))
}

func TestAllTrue(t *testing.T) {
	assert.False(t, AllTrue(false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false))
	assert.False(t, AllTrue(false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false))
	assert.True(t, AllTrue(true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true))
	assert.False(t, AllTrue(true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true))
}

func TestAnd(t *testing.T) {
	assert.False(t, And())
	assert.True(t, And(true))
	assert.False(t, And(false))
	assert.True(t, And(true, true))
	assert.False(t, And(false, false))
	assert.False(t, And(true, false))
	assert.False(t, And(true, true, false))
	assert.True(t, And(true, true, true))
}

func TestConditions_Add(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.Empty(t, emptyConditions.Add())
	conditions := NewConditionsFromValues(false, true, false)
	assert.EqualValues(t, conditions, conditions.Add())
	assert.EqualValues(t, []bool{false, true, false, false, true, false}, conditions.Add(false, true, false))
}

func TestConditions_All(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.False(t, emptyConditions.All())
	conditions := NewConditionsFromValues(false, false, false)
	assert.False(t, conditions.All())
	conditions = NewConditionsFromValues(false, true, false)
	assert.False(t, conditions.All())
	conditions = NewConditionsFromValues(false, true, true)
	assert.False(t, conditions.All())
	conditions = NewConditionsFromValues(true, true, true)
	assert.True(t, conditions.All())
}

func TestConditions_And(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.False(t, emptyConditions.And())
	conditions := NewConditionsFromValues(false, false, false)
	assert.False(t, conditions.And())
	conditions = NewConditionsFromValues(false, true, false)
	assert.False(t, conditions.And())
	conditions = NewConditionsFromValues(false, true, true)
	assert.False(t, conditions.And())
	conditions = NewConditionsFromValues(true, true, true)
	assert.True(t, conditions.All())
}

func TestConditions_Any(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.False(t, emptyConditions.Any())
	conditions := NewConditionsFromValues(false, false, false)
	assert.False(t, conditions.Any())
	conditions = NewConditionsFromValues(false, true, false)
	assert.True(t, conditions.Any())
	conditions = NewConditionsFromValues(false, true, true)
	assert.True(t, conditions.Any())
}

func TestConditions_Concat(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.Empty(t, emptyConditions.Concat(&emptyConditions))
	conditions := NewConditionsFromValues(false, true, false)
	assert.EqualValues(t, []bool{false, true, false, false, true, false}, conditions.Concat(&conditions))
}

func TestConditions_Contains(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.False(t, emptyConditions.Contains(true))
	assert.True(t, emptyConditions.Contains(false))
	conditions := NewConditionsFromValues(false, true, false)
	assert.True(t, conditions.Contains(true))
	assert.True(t, conditions.Contains(false))
	conditions = NewConditionsFromValues(false, false, false)
	assert.False(t, conditions.Contains(true))
	assert.True(t, conditions.Contains(false))
	conditions = NewConditionsFromValues(true, true, true)
	assert.True(t, conditions.Contains(true))
	assert.False(t, conditions.Contains(false))
}

func TestConditions_ForEach(t *testing.T) {
	count := 0
	increment := func(b bool) error {
		count++
		return nil
	}
	emptyConditions := NewConditionsFromValues()
	err := emptyConditions.ForEach(increment)
	errortest.AssertError(t, err, commonerrors.ErrUndefined)
	assert.Empty(t, count)
	conditions := NewConditionsFromValues(false, true, false)
	err = conditions.ForEach(increment)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestConditions_Negate(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.Empty(t, emptyConditions.Negate())
	conditions := NewConditionsFromValues(false, false, false)
	negatedConditions := conditions.Negate()
	assert.NotEqual(t, conditions, negatedConditions)
	assert.EqualValues(t, []bool{true, true, true}, negatedConditions)
}

func TestConditions_OneHot(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.False(t, emptyConditions.OneHot())
	conditions := NewConditionsFromValues(false, false, false)
	assert.False(t, conditions.OneHot())
	conditions = NewConditionsFromValues(false, true, false)
	assert.True(t, conditions.OneHot())
	conditions = NewConditionsFromValues(false, true, true)
	assert.False(t, conditions.OneHot())
}

func TestConditions_Or(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.False(t, emptyConditions.Or())
	conditions := NewConditionsFromValues(false, false, false)
	assert.False(t, conditions.Or())
	conditions = NewConditionsFromValues(false, true, false)
	assert.True(t, conditions.Or())
}

func TestConditions_Xor(t *testing.T) {
	emptyConditions := NewConditionsFromValues()
	assert.False(t, emptyConditions.Xor())
	oddConditions := NewConditionsFromValues(true, true, true)
	assert.True(t, oddConditions.Xor())
	evenConditions := NewConditionsFromValues(true, true, true, true)
	assert.False(t, evenConditions.Xor())
}

func TestNegate(t *testing.T) {
	assert.Empty(t, Negate())
	assert.EqualValues(t, []bool{false}, Negate(true))
	assert.EqualValues(t, []bool{true}, Negate(false))
	assert.EqualValues(t, []bool{true, false, true, false, false, true, true}, Negate(false, true, false, true, true, false, false))
}

func TestOneHot(t *testing.T) {
	assert.False(t, OneHot())
	assert.True(t, OneHot(true))
	assert.False(t, OneHot(false))
	assert.False(t, OneHot(true, true))
	assert.False(t, OneHot(false, false))
	assert.True(t, OneHot(true, false))
	assert.False(t, OneHot(false, false, false))
	assert.True(t, OneHot(true, false, false))
	assert.True(t, OneHot(false, true, false))
	assert.True(t, OneHot(false, false, true))
	assert.False(t, OneHot(false, true, true))
	assert.False(t, OneHot(true, false, true))
	assert.False(t, OneHot(true, true, false))
	assert.False(t, OneHot(true, true, true))

}

func TestOr(t *testing.T) {
	assert.False(t, Or())
	assert.True(t, Or(true))
	assert.False(t, Or(false))
	assert.True(t, Or(true, true))
	assert.False(t, Or(false, false))
	assert.True(t, Or(true, false))
	assert.True(t, Or(true, true, false))
	assert.True(t, Or(true, true, true))
	assert.False(t, Or(false, false, false))
}

func TestXor(t *testing.T) {
	assert.False(t, Xor())
	assert.True(t, Xor(true))
	assert.False(t, Xor(false))
	assert.False(t, Xor(true, true))
	assert.False(t, Xor(false, false))
	assert.True(t, Xor(true, false))
	assert.Equal(t, xor(xor(true, true), false), Xor(true, true, false))
	assert.False(t, Xor(true, true, false))
	assert.Equal(t, xor(xor(true, false), false), Xor(true, false, false))
	assert.True(t, Xor(true, false, false))
	assert.Equal(t, xor(xor(true, true), true), Xor(true, true, true))
	assert.True(t, Xor(true, true, true))
	assert.Equal(t, xor(xor(true, xor(true, true)), true), Xor(true, true, true, true))
	assert.False(t, Xor(true, true, true, true))

}
