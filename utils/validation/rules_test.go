package validation

import (
	"encoding/base64"
	"regexp"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestCastingToInt(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		err   error
	}{
		{"int", int(8080), nil},
		{"int8", int8(80), nil},
		{"int16", int16(8080), nil},
		{"int32", int32(8080), nil},
		{"int64", int64(8080), nil},
		{"uint", uint(8080), nil},
		{"uint8", uint8(80), nil},
		{"uint16", uint16(8080), nil},
		{"uint32", uint32(8080), nil},
		{"uint64", uint64(8080), nil},
		{"string valid", "8080", nil},
		{"[]byte valid", []byte("8080"), nil},
		{"int min valid port", int(1), nil},
		{"int max valid port", int(65535), nil},
		{"string min valid port", "1", nil},
		{"string max valid port", "65535", nil},
		{"int below range", int(0), commonerrors.ErrInvalid},
		{"int above range", int(65536), commonerrors.ErrInvalid},
		{"uint above range", uint(65536), commonerrors.ErrInvalid},
		{"string negative", "-1", commonerrors.ErrInvalid},
		{"string above range", "65536", commonerrors.ErrInvalid},
		{"string non-numeric", "notaport", commonerrors.ErrInvalid},
		{"[]byte non-numeric", []byte("notaport"), commonerrors.ErrInvalid},
		{"float64", float64(8080), commonerrors.ErrMarshalling},
		{"bool", true, commonerrors.ErrMarshalling},
		{"struct", struct{}{}, commonerrors.ErrMarshalling},
		{"nil", nil, commonerrors.ErrMarshalling},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := IsPort.Validate(test.value)
			if test.err == nil {
				assert.NoError(t, err)
			} else {
				errortest.AssertError(t, err, test.err)
			}
		})
	}
}

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
		{"", true},                     // Empty string
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
			err := validation.Validate(test.input, IsBase64)
			if test.expected {
				require.NoError(t, err)
			} else {
				errortest.AssertErrorDescription(t, err, is.ErrBase64.Error())
			}
		})
	}
}

func TestIsPathParameter(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"{abc}", true},
		{"abc}", false},
		{"{abc", false},
		{"abc", false},
		{"{abc$123.zzz~999}", true},
		{"{abc%5F1}", true}, // unescaped as '{abc_1}'
		{"{abc#123}", false},
		{" ", false},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.input, func(t *testing.T) {
			err := validation.Validate(test.input, IsPathParameter)
			if test.expected {
				require.NoError(t, err)
			} else {
				errortest.AssertErrorDescription(t, err, "invalid path parameter")
			}
		})
	}
}

func TestAdditionalGeneralRules(t *testing.T) {
	t.Run("string rules", func(t *testing.T) {
		assert.NoError(t, validation.Validate("prefix-value", Prefix("pre")))
		assert.Error(t, validation.Validate("value", Prefix("pre")))

		assert.NoError(t, validation.Validate("file.txt", Suffix(".txt")))
		assert.Error(t, validation.Validate("file.log", Suffix(".txt")))

		assert.NoError(t, validation.Validate("hello world", ContainsString("world")))
		assert.Error(t, validation.Validate("hello", ContainsString("world")))

		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate("hello", Like(re)))
		assert.Error(t, validation.Validate("123", Like(re)))

		assert.NoError(t, validation.Validate("hello", NotContains("world")))
		assert.Error(t, validation.Validate("hello world", NotContains("world")))

		assert.NoError(t, validation.Validate("helloworld", NotContainsWhitespaces()))
		assert.Error(t, validation.Validate("hello world", NotContainsWhitespaces()))
	})

	t.Run("length and occurs", func(t *testing.T) {
		assert.NoError(t, validation.Validate("ab", LengthExact(2)))
		assert.Error(t, validation.Validate("abc", LengthExact(2)))

		assert.NoError(t, validation.Validate("banana", MinOccurs("a", 3)))
		assert.Error(t, validation.Validate("banana", MinOccurs("a", 4)))

		assert.NoError(t, validation.Validate("banana", MaxOccurs("a", 3)))
		assert.Error(t, validation.Validate("banana", MaxOccurs("a", 2)))

		assert.NoError(t, validation.Validate("banana", OccursExactly("a", 3)))
		assert.Error(t, validation.Validate("banana", OccursExactly("a", 2)))
	})

	t.Run("time rules", func(t *testing.T) {
		assert.NoError(t, validation.Validate("5s", IsDuration))
		assert.Error(t, validation.Validate("not-a-duration", IsDuration))
		assert.NoError(t, validation.Validate("10s", DurationMinimum(5*time.Second)))
		assert.Error(t, validation.Validate("1s", DurationMinimum(5*time.Second)))
		assert.NoError(t, validation.Validate("5s", DurationConst(5*time.Second)))
		assert.Error(t, validation.Validate("6s", DurationConst(5*time.Second)))

		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", IsTimestamp))
		assert.Error(t, validation.Validate("2024-01-01", IsTimestamp))
		minTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		maxTime := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMinimum(minTime)))
		assert.Error(t, validation.Validate("2023-12-31T00:00:00Z", TimestampMinimum(minTime)))
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMaximum(maxTime)))
		assert.Error(t, validation.Validate("2025-01-01T00:00:00Z", TimestampMaximum(maxTime)))
		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", TimestampConst(minTime)))
		assert.Error(t, validation.Validate("2024-01-02T00:00:00Z", TimestampConst(minTime)))
	})
}
