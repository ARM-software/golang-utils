package validation

import (
	"context"
	"iter"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

type testRuleWithContext func(context.Context, any) error

func (r testRuleWithContext) ValidateWithContext(ctx context.Context, value any) error {
	return r(ctx, value)
}

func wrapRuleWithContext(rule validation.Rule) validation.RuleWithContext {
	return testRuleWithContext(func(_ context.Context, value any) error {
		return rule.Validate(value)
	})
}

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

// Some sequences are consumed after one iteration. Composite rules will break it. You mention it on utils/validation/utils.go:135 but don't consider it here
func TestCompositeRulesWithChannelSequence(t *testing.T) {
	items := make(chan any, 1)
	items <- "not-an-integer"
	close(items)

	sequence := iter.Seq[any](func(yield func(any) bool) {
		for item := range items {
			if !yield(item) {
				return
			}
		}
	})

	err := AllOf(MinItems(1), ArrayItems(Type("integer"))).Validate(sequence)

	require.Error(t, err)
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
		require.NoError(t, IfThenElse(nil, validation.Required, is.Email).Validate("value"))
		errortest.AssertError(t, IfThenElse(nil, validation.Required, is.Email).Validate(""), commonerrors.ErrInvalid)
		require.NoError(t, IfThenElse(nil, nil, is.Email).Validate("plain-text"))
		require.NoError(t, IfThenElse(nil, nil, is.Email).Validate("user@example.com"))
		require.NoError(t, IfThenElse(is.Email, nil, nil).Validate("user@example.com"))
		require.NoError(t, IfThenElse(is.Email, nil, nil).Validate("plain-text"))
		require.NoError(t, IfThenElse(nil, nil, nil).Validate("anything"))
	})
}

// context cancellation returns a validation error not a cancelled
func TestAtMostWithContextPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rule := testRuleWithContext(func(ctx context.Context, _ any) error {
		return ctx.Err()
	})

	err := AtMostWithContext(0, rule).ValidateWithContext(ctx, "value")

	require.ErrorIs(t, err, context.Canceled)
}

func TestContextualRules(t *testing.T) {
	ctx := context.Background()
	email := wrapRuleWithContext(is.Email)
	uuid := wrapRuleWithContext(is.UUID)
	required := wrapRuleWithContext(validation.Required)

	t.Run("append contextual rule", func(t *testing.T) {
		rule := NewAnyCompositeRule()
		rule.AppendContextualRule(email)

		require.NoError(t, rule.ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, rule.ValidateWithContext(ctx, "plain-text"), commonerrors.ErrInvalid)
	})

	t.Run("new any rule with context", func(t *testing.T) {
		require.NoError(t, NewAnyRuleWithContext(email, uuid).ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, NewAnyRuleWithContext(email, uuid).ValidateWithContext(ctx, "plain-text"), commonerrors.ErrInvalid)
	})

	t.Run("new all rule with context", func(t *testing.T) {
		require.NoError(t, NewAllRuleWithContext(required, email).ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, NewAllRuleWithContext(required, email).ValidateWithContext(ctx, "plain-text"), commonerrors.ErrInvalid)
	})

	t.Run("new none rule with context", func(t *testing.T) {
		require.NoError(t, NewNoneRuleWithContext(email, uuid).ValidateWithContext(ctx, "plain-text"))
		errortest.AssertError(t, NewNoneRuleWithContext(email, uuid).ValidateWithContext(ctx, "user@example.com"), commonerrors.ErrInvalid)
	})

	t.Run("at most with context", func(t *testing.T) {
		require.NoError(t, AtMostWithContext(1, email, uuid).ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, AtMostWithContext(1, required, email).ValidateWithContext(ctx, "user@example.com"), commonerrors.ErrInvalid)
	})

	t.Run("exactly with context", func(t *testing.T) {
		require.NoError(t, ExactlyWithContext(1, email, uuid).ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, ExactlyWithContext(1, required, email).ValidateWithContext(ctx, "user@example.com"), commonerrors.ErrInvalid)
	})

	t.Run("nof with context", func(t *testing.T) {
		require.NoError(t, NOfWithContext(1, email, uuid).ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, NOfWithContext(2, email, uuid).ValidateWithContext(ctx, "user@example.com"), commonerrors.ErrInvalid)
	})

	t.Run("at least with context", func(t *testing.T) {
		require.NoError(t, AtLeastWithContext(1, email, uuid).ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, AtLeastWithContext(2, email, uuid).ValidateWithContext(ctx, "user@example.com"), commonerrors.ErrInvalid)
	})

	t.Run("implies with context", func(t *testing.T) {
		require.NoError(t, ImpliesWithContext(email, required).ValidateWithContext(ctx, "user@example.com"))
		require.NoError(t, ImpliesWithContext(email, required).ValidateWithContext(ctx, "plain-text"))
		errortest.AssertError(t, ImpliesWithContext(required, email).ValidateWithContext(ctx, "plain-text"), commonerrors.ErrInvalid)
		require.NoError(t, ImpliesWithContext(nil, email).ValidateWithContext(ctx, "plain-text"))
		require.NoError(t, ImpliesWithContext(email, nil).ValidateWithContext(ctx, "user@example.com"))
	})

	t.Run("if then else with context", func(t *testing.T) {
		require.NoError(t, IfThenElseWithContext(email, required, nil).ValidateWithContext(ctx, "user@example.com"))
		require.NoError(t, IfThenElseWithContext(email, nil, required).ValidateWithContext(ctx, "plain-text"))
		errortest.AssertError(t, IfThenElseWithContext(email, uuid, nil).ValidateWithContext(ctx, "user@example.com"), commonerrors.ErrInvalid)
		errortest.AssertError(t, IfThenElseWithContext(email, nil, email).ValidateWithContext(ctx, "plain-text"), commonerrors.ErrInvalid)
		require.NoError(t, IfThenElseWithContext(nil, required, email).ValidateWithContext(ctx, "value"))
	})

	t.Run("one of with context", func(t *testing.T) {
		require.NoError(t, NewOneOfRuleWithContext(email, uuid).ValidateWithContext(ctx, "user@example.com"))
		errortest.AssertError(t, NewOneOfRuleWithContext(required, email).ValidateWithContext(ctx, "user@example.com"), commonerrors.ErrInvalid)
	})
}
