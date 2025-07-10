package base64

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestIsBase64Encoded(t *testing.T) {
	random := faker.Sentence()
	base641 := base64.RawURLEncoding.EncodeToString([]byte(random))
	base642 := base64.RawStdEncoding.EncodeToString([]byte(random))
	base643 := base64.URLEncoding.EncodeToString([]byte(random))
	base644 := base64.StdEncoding.EncodeToString([]byte(random))
	tests := []struct {
		input    string
		expected bool
	}{
		{"U29tZSBkYXRh", true},     // "Some data"
		{"SGVsbG8gd29ybGQ=", true}, // "Hello world"
		{"U29tZSBkYXRh===", false},
		{"", false},                    // Empty string
		{"NotBase64", false},           // Plain text
		{"!@#$%^&*", false},            // Non-Base64 characters
		{"U29tZSBkYXRh\n", true},       // Line break
		{"V2l0aCB3aGl0ZXNwYWNl", true}, // "With whitespace" (valid if stripped)
		{base641, true},
		{base642, true},
		{base643, true},
		{base644, true},
		{"U29tZSBkYXRh=", true},
		{"U29tZSBkYXRh==", true},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.input, func(t *testing.T) {
			if test.expected {
				assert.True(t, IsEncoded(test.input))
			} else {
				assert.False(t, IsEncoded(test.input))
			}
		})
	}
}

func TestDecodeIfBase64(t *testing.T) {
	random := faker.Sentence()
	base641 := base64.RawURLEncoding.EncodeToString([]byte(random))
	base642 := base64.RawStdEncoding.EncodeToString([]byte(random))
	base643 := base64.URLEncoding.EncodeToString([]byte(random))
	base644 := base64.StdEncoding.EncodeToString([]byte(random))

	tests := []struct {
		input    string
		expected string
		errors   bool
	}{
		{input: "U29tZSBkYXRh", expected: "Some data"},
		{input: "SGVsbG8gd29ybGQ=", expected: "Hello world"},
		{input: "VGVzdCBzdHJpbmc=", expected: "Test string"},
		{input: "MTIzNDU2", expected: "123456"},
		{input: base641, expected: random},
		{input: base642, expected: random},
		{input: base643, expected: random},
		{input: base644, expected: random},

		{input: "NotBase64", expected: "NotBase64", errors: true},
		{input: "Invalid===", expected: "Invalid===", errors: true},
		{input: "", expected: "", errors: true},
		{input: "!@#$%^&*", expected: "!@#$%^&*", errors: true},

		{input: "U29tZSBkYXRh\n", expected: "Some data"}, // newline is not part of valid base64
		{input: "U29tZSBkYXRh=", expected: "Some data"},  // valid with single padding
		{input: "U29tZSBkYXRh==", expected: "Some data"}, // valid with double padding
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.input, func(t *testing.T) {
			result, err := DecodeString(context.Background(), test.input)
			assert.Equal(t, test.expected, DecodeIfEncoded(context.Background(), test.input))
			if test.errors {
				errortest.AssertError(t, err, commonerrors.ErrMarshalling, commonerrors.ErrInvalid, commonerrors.ErrEmpty)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}

	t.Run("cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := DecodeString(ctx, random)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.Equal(t, random, DecodeIfEncoded(ctx, random))

	})
}

func TestDecodeRecursively(t *testing.T) {
	randomText := faker.Paragraph()
	random, err := faker.RandomInt(1, 10, 1)
	require.NoError(t, err)

	encodedText := randomText
	for i := 0; i < random[0]; i++ {
		encodedText = EncodeString(encodedText)
	}

	assert.NotEqual(t, randomText, encodedText)
	assert.Equal(t, randomText, DecodeRecursively(context.Background(), encodedText))
}
