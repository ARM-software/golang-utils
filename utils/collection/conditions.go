package collection

import (
	"iter"
	"slices"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type Conditions []bool

// NewConditions creates a set of conditions.
func NewConditions(capacity int) Conditions {
	return make([]bool, 0, capacity)
}

// NewConditionsFromValues creates a set of conditions.
func NewConditionsFromValues(conditions ...bool) Conditions {
	c := NewConditions(len(conditions))
	c.Add(conditions...)
	return c
}

// Add adds conditions and returns itself.
func (c *Conditions) Add(conditions ...bool) Conditions {
	if c == nil {
		return nil
	}
	*c = append(*c, conditions...)
	return *c
}

// ForEach will execute function each() on every condition unless an error is returned and will end add this point.
func (c *Conditions) ForEach(each func(bool) error) error {
	if c == nil || len(*c) == 0 {
		return commonerrors.New(commonerrors.ErrUndefined, "the collection of conditions is empty")
	}
	for i := range *c {
		subErr := each((*c)[i])
		if subErr != nil {
			return subErr
		}
	}
	return nil
}

// Contains returns whether the conditions collection contains the value of a condition.
func (c *Conditions) Contains(condition bool) bool {
	if c == nil {
		return false
	}
	if condition {
		return c.Any()
	} else {
		return !c.All()
	}
}

// All returns true if all conditions are true.
func (c *Conditions) All() bool {
	if c == nil {
		return false
	}
	return All(*c)
}

// Any returns true if  there is at least one condition which is true.
func (c *Conditions) Any() bool {
	if c == nil {
		return false
	}
	return Any(*c)
}

// Concat concatenates conditions and returns itself.
func (c *Conditions) Concat(more *Conditions) Conditions {
	if more == nil {
		return nil
	}
	return c.Add(*more...)
}

// Negate returns a new set of conditions with negated values.
func (c *Conditions) Negate() Conditions {
	if c == nil {
		return nil
	}
	negation := Conditions(Negate(*c...))
	return negation
}

// And performs an `and` operation on all conditions
func (c *Conditions) And() bool {
	if c == nil {
		return false
	}
	return And(*c...)
}

// Or performs an `Or` operation on all conditions
func (c *Conditions) Or() bool {
	if c == nil {
		return false
	}
	return Or(*c...)
}

// Xor performs a `Xor` operation on all conditions
func (c *Conditions) Xor() bool {
	if c == nil {
		return false
	}
	return Xor(*c...)
}

// OneHot performs an `OneHot` operation on all conditions
func (c *Conditions) OneHot() bool {
	if c == nil {
		return false
	}
	return OneHot(*c...)
}

// Any returns true if there is at least one element of the slice which is true.
func Any(slice []bool) bool {
	return AnySequence(slices.Values(slice))
}

// AnySequence returns true if there is at least one element of the slice which is true.
func AnySequence(seq iter.Seq[bool]) bool {
	if seq == nil {
		return false
	}
	for e := range seq {
		if e {
			return true
		}
	}
	return false
}

// AnyTrue returns whether there is a value set to true
func AnyTrue(values ...bool) bool {
	return Any(values)
}

// AnyFalseSequence returns true if there is at least one element of the sequence which is false.
func AnyFalseSequence(eq iter.Seq[bool]) bool {
	hasElements := false
	for e := range eq {
		hasElements = true
		if !e {
			return true
		}
	}
	return !hasElements
}

// AnyFalse returns whether there is a value set to false
func AnyFalse(values ...bool) bool {
	return AnyFalseSequence(slices.Values(values))
}

// AllSequence returns true if all items of the sequence are true.
func AllSequence(seq iter.Seq[bool]) bool {
	return !AnyFalseSequence(seq)
}

// All returns true if all items of the slice are true.
func All(slice []bool) bool {
	return AllSequence(slices.Values(slice))
}

// AllTrue returns whether all values are true.
func AllTrue(values ...bool) bool {
	return All(values)
}

// Negate returns the slice with contrary values.
func Negate(values ...bool) []bool {
	if values == nil {
		return nil
	}
	if len(values) == 0 {
		return []bool{}
	}
	negatedValues := make([]bool, len(values))
	for i := range values {
		negatedValues[i] = !values[i]
	}
	return negatedValues
}

// And performs an 'and' operation on an array of booleans.
func And(values ...bool) bool {
	return All(values)
}

// Or performs an 'or' operation on an array of booleans.
func Or(values ...bool) bool {
	return Any(values)
}

// Xor performs a `xor` on an array of booleans. This behaves like an XOR gate; it returns true if the number of true values is odd, and false if the number of true values is zero or even.
func Xor(values ...bool) bool {
	if len(values) == 0 {
		return false
	}
	// false if the neutral element of the xor operator
	result := false
	for i := range values {
		result = xor(result, values[i])
	}
	return result
}

// xor(true, true)   = false
// xor(false, false) = false
// xor(true, false)  = true
func xor(a, b bool) bool {
	return a != b
}

// OneHot returns true if one, and only one, of the supplied values is true. See https://en.wikipedia.org/wiki/One-hot
func OneHot(values ...bool) bool {
	if len(values) == 0 {
		return false
	}
	count := 0
	for i := range values {
		if values[i] {
			count++
		}
	}
	return count == 1
}
