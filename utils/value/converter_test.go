package value

import (
	"context"
	"errors"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stringValue string

func (s stringValue) String() string {
	return "stringer:" + string(s)
}

type textValue string

func (t textValue) MarshalText() ([]byte, error) {
	return []byte("text:" + string(t)), nil
}

func (t textValue) String() string {
	return "stringer:" + string(t)
}

type failingTextValue struct{}

func (failingTextValue) MarshalText() ([]byte, error) {
	return nil, errors.New("boom")
}

type marshalOnlyValue string

func (m marshalOnlyValue) MarshalText() ([]byte, error) {
	return []byte("text:" + string(m)), nil
}

type nilStringer struct{}

func (*nilStringer) String() string {
	return "unexpected"
}

func TestIdentityConverter(t *testing.T) {
	value, err := IdentityConverter.ConvertValue(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestStringConverter(t *testing.T) {
	t.Run("fallback format", func(t *testing.T) {
		value, err := StringConverter.ConvertValue(context.Background(), 42)
		require.NoError(t, err)
		assert.Equal(t, "42", value)
	})

	t.Run("string value", func(t *testing.T) {
		expected := faker.Sentence()
		value, err := StringConverter.ConvertValue(context.Background(), expected)
		require.NoError(t, err)
		assert.Equal(t, expected, value)
	})

	t.Run("stringer", func(t *testing.T) {
		value, err := StringConverter.ConvertValue(context.Background(), stringValue("value"))
		require.NoError(t, err)
		assert.Equal(t, "stringer:value", value)
	})

	t.Run("stringer preferred over text marshaler", func(t *testing.T) {
		value, err := StringConverter.ConvertValue(context.Background(), textValue("value"))
		require.NoError(t, err)
		assert.Equal(t, "stringer:value", value)
	})

	t.Run("text marshaler", func(t *testing.T) {
		value, err := StringConverter.ConvertValue(context.Background(), marshalOnlyValue("value"))
		require.NoError(t, err)
		assert.Equal(t, "text:value", value)
	})

	t.Run("marshal text error", func(t *testing.T) {
		_, err := StringConverter.ConvertValue(context.Background(), failingTextValue{})
		require.Error(t, err)
		assert.EqualError(t, err, "boom")
	})

	t.Run("nil stringer pointer falls back", func(t *testing.T) {
		var value *nilStringer
		converted, err := StringConverter.ConvertValue(context.Background(), value)
		require.NoError(t, err)
		assert.Equal(t, "<nil>", converted)
	})
}
