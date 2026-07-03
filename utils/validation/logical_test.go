package validation

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestCompositeRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		rule  validation.Rule
		valid bool
	}{
		{
			value: "user@example.com",
			rule: NewAnyRule(
				is.Email,
				is.UUID,
			),
			valid: true,
		},
		{
			value: "550e8400-e29b-41d4-a716-446655440000",
			rule: NewAnyRule(
				is.Email,
				is.UUID,
			),
			valid: true,
		},
		{
			value: "not-valid",
			rule: NewAnyRule(
				is.Email,
				is.UUID,
			),
			valid: false,
		},
		{
			value: "https://example.com",
			rule: NewAnyRule(
				is.Email,
				is.UUID,
				is.URL,
			),
			valid: true,
		},
		{
			value: "",
			rule: NewAnyRule(
				is.Email,
				is.UUID,
				is.URL,
			),
			valid: true,
		},
		{
			value: "",
			rule: NewAnyRule(
				validation.Required,
			),
			valid: false,
		},
		{
			value: "",
			rule: NewAllRule(
				validation.Required,
				is.Email,
				is.UUID,
				is.URL,
			),
			valid: false,
		},
		{
			value: "",
			rule: NewNoneRule(
				is.Email,
				is.URL,
			),
			valid: false,
		},
		{
			value: "",
			rule: NewAllRule(
				validation.Required,
				is.Email,
				is.UUID,
				is.URL,
			),
			valid: false,
		},
		{
			value: "user@example.com",
			rule: NewAllRule(
				validation.Required,
				is.Email,
			),
			valid: true,
		},
		{
			value: "user@example.com",
			rule: NewAllRule(
				validation.Required,
				is.Email,
				is.UUID,
			),
			valid: false,
		},
		{
			value: "",
			rule: NewNoneRule(
				is.Email,
				is.UUID,
				is.URL,
			),
			valid: false,
		},
		{
			value: "user@example.com",
			rule: NewNoneRule(
				is.Email,
				is.UUID,
			),
			valid: false,
		},
		{
			value: "plain-text",
			rule: NewNoneRule(
				is.Email,
				is.UUID,
			),
			valid: true,
		},
		{
			value: "user@example.com",
			rule: NewNoneRule(
				is.Email,
			),
			valid: false,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.value, func(t *testing.T) {
			t.Parallel()

			err := test.rule.Validate(test.value)

			if test.valid {
				require.NoError(t, err)
			} else {
				errortest.AssertError(t, err, commonerrors.ErrInvalid)
			}
			err = validation.Validate(test.value, test.rule)
			if test.valid {
				require.NoError(t, err)
			} else {
				errortest.AssertError(t, err, commonerrors.ErrInvalid)
			}
		})
	}
}

func TestCountingRules(t *testing.T) {
	t.Run("at least", func(t *testing.T) {
		require.NoError(t, AtLeast(1, is.Email, is.UUID).Validate("user@example.com"))
		errortest.AssertError(t, AtLeast(2, is.Email, is.UUID).Validate("user@example.com"), commonerrors.ErrInvalid)
		require.NoError(t, AtLeast(0, is.Email, is.UUID).Validate("plain-text"))
	})

	t.Run("nof", func(t *testing.T) {
		require.NoError(t, NOf(1, is.Email, is.UUID).Validate("user@example.com"))
		errortest.AssertError(t, NOf(1, validation.Required, is.Email).Validate("user@example.com"), commonerrors.ErrInvalid)
		errortest.AssertError(t, NOf(2, is.Email, is.UUID).Validate("user@example.com"), commonerrors.ErrInvalid)
	})

	t.Run("at most", func(t *testing.T) {
		require.NoError(t, AtMost(1, is.Email, is.UUID).Validate("user@example.com"))
		errortest.AssertError(t, AtMost(1, validation.Required, is.Email).Validate("user@example.com"), commonerrors.ErrInvalid)
		require.NoError(t, AtMost(0, is.Email, is.UUID).Validate("plain-text"))
	})

	t.Run("exactly", func(t *testing.T) {
		require.NoError(t, Exactly(1, is.Email, is.UUID).Validate("user@example.com"))
		errortest.AssertError(t, Exactly(1, validation.Required, is.Email).Validate("user@example.com"), commonerrors.ErrInvalid)
		errortest.AssertError(t, Exactly(1, is.Email, is.UUID).Validate("plain-text"), commonerrors.ErrInvalid)
	})

	t.Run("implies", func(t *testing.T) {
		require.NoError(t, Implies(is.Email, validation.Required).Validate("user@example.com"))
		require.NoError(t, Implies(is.Email, validation.Required).Validate("plain-text"))
		errortest.AssertError(t, Implies(validation.Required, is.Email).Validate("plain-text"), commonerrors.ErrInvalid)
		require.NoError(t, Implies(nil, is.Email).Validate("plain-text"))
		require.NoError(t, Implies(is.Email, nil).Validate("user@example.com"))
	})

	t.Run("if then else", func(t *testing.T) {
		require.NoError(t, IfThenElse(is.Email, validation.Required, nil).Validate("user@example.com"))
		require.NoError(t, IfThenElse(is.Email, nil, validation.Required).Validate("plain-text"))
		errortest.AssertError(t, IfThenElse(is.Email, is.UUID, nil).Validate("user@example.com"), commonerrors.ErrInvalid)
		errortest.AssertError(t, IfThenElse(is.Email, nil, is.Email).Validate("plain-text"), commonerrors.ErrInvalid)
	})
}
