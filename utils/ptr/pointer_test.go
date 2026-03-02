package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTo(t *testing.T) {
	t.Run("returns_non_nil_pointer_with_expected_value", func(t *testing.T) {
		p := To(123)
		require.NotNil(t, p)
		assert.Equal(t, 123, *p)
	})

	t.Run("returns_pointer_to_copy_not_alias", func(t *testing.T) {
		original := 41

		p := To(original)
		require.NotNil(t, p)
		assert.Equal(t, 41, *p)

		*p = 99
		assert.Equal(t, 41, original, "mutating the returned pointer must not change the original value")
	})
}

func TestToOrNilIfEmpty(t *testing.T) {
	t.Run("empty_string_returns_nil", func(t *testing.T) {
		got := ToOrNilIfEmpty("")
		require.Nil(t, got)
	})

	t.Run("whitespace_only_string_returns_nil", func(t *testing.T) {
		got := ToOrNilIfEmpty("   \t ")
		require.Nil(t, got)
	})

	t.Run("non_empty_string_returns_pointer", func(t *testing.T) {
		got := ToOrNilIfEmpty("hello")
		require.NotNil(t, got)
		assert.Equal(t, "hello", *got)
	})

	t.Run("zero_int_returns_nil", func(t *testing.T) {
		got := ToOrNilIfEmpty(0)
		require.Nil(t, got)
	})

	t.Run("non_zero_int_returns_pointer", func(t *testing.T) {
		got := ToOrNilIfEmpty(7)
		require.NotNil(t, got)
		assert.Equal(t, 7, *got)
	})

	t.Run("false_returns_nil", func(t *testing.T) {
		got := ToOrNilIfEmpty(false)
		require.Nil(t, got)
	})

	t.Run("true_returns_pointer", func(t *testing.T) {
		got := ToOrNilIfEmpty(true)
		require.NotNil(t, got)
		assert.Equal(t, true, *got)
	})

	t.Run("empty_slice_returns_nil", func(t *testing.T) {
		got := ToOrNilIfEmpty([]int{})
		require.Nil(t, got)
	})

	t.Run("non_empty_slice_returns_pointer", func(t *testing.T) {
		in := []int{1, 2, 3}
		got := ToOrNilIfEmpty(in)
		require.NotNil(t, got)
		assert.Equal(t, in, *got)
	})

	t.Run("nil_pointer_value_returns_nil", func(t *testing.T) {
		var s *string = nil
		got := ToOrNilIfEmpty(s)
		require.Nil(t, got)
	})
}

func TestFrom(t *testing.T) {
	t.Run("nil_pointer_returns_zero_value", func(t *testing.T) {
		var p *int = nil
		assert.Equal(t, 0, From(p))
	})

	t.Run("non_nil_pointer_returns_value", func(t *testing.T) {
		v := 123
		assert.Equal(t, 123, From(&v))
	})
}

func TestFromOrDefault(t *testing.T) {
	t.Run("nil_pointer_returns_default", func(t *testing.T) {
		var p *string = nil
		assert.Equal(t, "default", FromOrDefault(p, "default"))
	})

	t.Run("non_nil_pointer_returns_value", func(t *testing.T) {
		v := "value"
		assert.Equal(t, "value", FromOrDefault(&v, "default"))
	})
}

func TestDeref(t *testing.T) {
	t.Run("matches_From_for_nil", func(t *testing.T) {
		var p *int = nil
		assert.Equal(t, From(p), Deref(p))
	})

	t.Run("matches_From_for_non_nil", func(t *testing.T) {
		v := 5
		assert.Equal(t, From(&v), Deref(&v))
	})
}
