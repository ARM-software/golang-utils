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
			valid: true,
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
