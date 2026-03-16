package validation

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// ICompositeRule represents a mutable logical rule set.
//
// A ICompositeRule can combine both standard ozzo rules and context-aware
// rules, and can itself be used anywhere a validation.Rule or
// validation.RuleWithContext is accepted, including with validation.Validate.
type ICompositeRule interface {
	// AppendRule adds one or more non-contextual rules to the composite rule.
	AppendRule(rule ...validation.Rule)

	// AppendContextualRule adds one or more context-aware rules to the
	// composite rule.
	AppendContextualRule(rule ...validation.RuleWithContext)

	validation.Rule
	validation.RuleWithContext
}

// compositeRule is the shared implementation for logical rule sets such as
// Any, All, and None.
type compositeRule struct {
	rules            []validation.Rule
	rulesWithContext []validation.RuleWithContext
}

// AppendRule adds one or more non-contextual rules to the composite rule.
func (r *compositeRule) AppendRule(rule ...validation.Rule) {
	r.rules = append(r.rules, rule...)
}

// AppendContextualRule adds one or more context-aware rules to the
// composite rule.
func (r *compositeRule) AppendContextualRule(rule ...validation.RuleWithContext) {
	r.rulesWithContext = append(r.rulesWithContext, rule...)
}

// verify evaluates all configured rules against v and returns one error per
// rule, preserving nil entries for rules that passed validation.
//
// Context-aware rules are evaluated with the provided context. Non-contextual
// rules are evaluated only while the context remains valid.
func (r *compositeRule) verify(context context.Context, v any) (verification []error) {
	verification = collection.Map[validation.RuleWithContext, error](r.rulesWithContext, func(sr validation.RuleWithContext) error {
		return sr.ValidateWithContext(context, v)
	})
	verification = append(verification, collection.Map[validation.Rule, error](r.rules, func(sr validation.Rule) error {
		err := parallelisation.DetermineContextError(context)
		if err != nil {
			return err
		}
		return sr.Validate(v)
	})...)
	return
}

// summariseErrors joins all non-nil validation errors into a single error.
func summariseErrors(errors []error) error {
	return commonerrors.Join(collection.Filter(errors, func(err error) bool { return err != nil })...)
}

// anyRule succeeds if at least one nested rule succeeds.
type anyRule struct {
	compositeRule
}

// Validate applies OR semantics to the configured rules and returns nil if at
// least one rule passes.
//
// This differs from validation.Validate, which applies AND semantics across
// the supplied rules.
func (r *anyRule) Validate(v any) error {
	errors := r.verify(context.Background(), v)
	return r.checkAny(errors)
}

// checkAny returns an invalid-value error if none of the evaluated rules
// succeeded.
func (r *anyRule) checkAny(errors []error) error {
	if !collection.AnyFunc(errors, func(err error) bool {
		return err == nil
	}) {
		return commonerrors.WrapError(commonerrors.ErrInvalid, summariseErrors(errors), "invalid value")
	}
	return nil
}

// ValidateWithContext applies OR semantics to the configured rules using ctx
// and returns nil if at least one rule passes.
func (r *anyRule) ValidateWithContext(ctx context.Context, v any) error {
	errors := r.verify(ctx, v)
	return r.checkAny(errors)
}

// NewAnyCompositeRule returns an empty composite rule with OR semantics.
//
// The returned rule succeeds if at least one appended rule succeeds, unlike
// validation.Validate which requires all supplied rules to succeed.
func NewAnyCompositeRule() ICompositeRule {
	return &anyRule{}
}

// NewAnyRule returns a rule that succeeds if at least one of the provided
// rules succeeds.
//
// This complements validation.Validate, where multiple rules are normally
// combined with AND semantics.
func NewAnyRule(rule ...validation.Rule) validation.Rule {
	c := NewAnyCompositeRule()
	c.AppendRule(rule...)
	return c
}

// NewAnyRuleWithContext returns a context-aware rule that succeeds if at
// least one of the provided context-aware rules succeeds.
//
// This complements validation.Validate and validation.ValidateWithContext,
// where multiple rules are normally combined with AND semantics.
func NewAnyRuleWithContext(rule ...validation.RuleWithContext) validation.RuleWithContext {
	c := NewAnyCompositeRule()
	c.AppendContextualRule(rule...)
	return c
}

// allRule succeeds only if all nested rules succeed.
type allRule struct {
	compositeRule
}

// Validate applies AND semantics to the configured rules and returns nil only
// if all rules pass.
//
// This matches the behaviour of validation.Validate when multiple rules are
// supplied.
func (r *allRule) Validate(v any) error {
	errors := r.verify(context.Background(), v)
	return r.checkAll(errors)
}

// checkAll returns an invalid-value error if any evaluated rule failed.
func (r *allRule) checkAll(errors []error) error {
	if !collection.AllFunc(errors, func(err error) bool {
		return err == nil
	}) {
		return commonerrors.WrapError(commonerrors.ErrInvalid, summariseErrors(errors), "invalid value")
	}
	return nil
}

// ValidateWithContext applies AND semantics to the configured rules using ctx
// and returns nil only if all rules pass.
func (r *allRule) ValidateWithContext(ctx context.Context, v any) error {
	errors := r.verify(ctx, v)
	return r.checkAll(errors)
}

// NewAllCompositeRule returns an empty composite rule with AND semantics.
//
// The returned rule succeeds only if all appended rules succeed, matching the
// behaviour of validation.Validate.
func NewAllCompositeRule() ICompositeRule {
	return &allRule{}
}

// NewAllRule returns a rule that succeeds only if all of the provided rules
// succeed.
//
// This is equivalent to grouping the same rules under validation.Validate.
func NewAllRule(rule ...validation.Rule) validation.Rule {
	c := NewAllCompositeRule()
	c.AppendRule(rule...)
	return c
}

// NewAllRuleWithContext returns a context-aware rule that succeeds only if
// all of the provided context-aware rules succeed.
//
// This is equivalent to grouping the same rules under
// validation.ValidateWithContext.
func NewAllRuleWithContext(rule ...validation.RuleWithContext) validation.RuleWithContext {
	c := NewAllCompositeRule()
	c.AppendContextualRule(rule...)
	return c
}

// noneRule succeeds only if none of the nested rules succeed.
type noneRule struct {
	compositeRule
}

// Validate applies NONE semantics to the configured rules and returns nil only
// if none of the rules pass.
//
// This provides the inverse of the usual validation.Validate behaviour.
func (r *noneRule) Validate(v any) error {
	errors := r.verify(context.Background(), v)
	return r.checkNone(errors)
}

// checkNone returns an invalid-value error if all evaluated rules succeeded.
func (r *noneRule) checkNone(errors []error) error {
	if collection.AllFunc(errors, func(err error) bool {
		return err == nil
	}) {
		return commonerrors.New(commonerrors.ErrInvalid, "invalid value")
	}
	return nil
}

// ValidateWithContext applies NONE semantics to the configured rules using ctx
// and returns nil only if none of the rules pass.
func (r *noneRule) ValidateWithContext(ctx context.Context, v any) error {
	errors := r.verify(ctx, v)
	return r.checkNone(errors)
}

// NewNoneCompositeRule returns an empty composite rule with NONE semantics.
//
// The returned rule succeeds only if none of the appended rules succeed.
func NewNoneCompositeRule() ICompositeRule {
	return &noneRule{}
}

// NewNoneRule returns a rule that succeeds only if none of the provided rules
// succeed.
//
// This is useful when a value must not match any rule in a given set.
func NewNoneRule(rule ...validation.Rule) validation.Rule {
	c := NewNoneCompositeRule()
	c.AppendRule(rule...)
	return c
}

// NewNoneRuleWithContext returns a context-aware rule that succeeds only if
// none of the provided context-aware rules succeed.
func NewNoneRuleWithContext(rule ...validation.RuleWithContext) validation.RuleWithContext {
	c := NewNoneCompositeRule()
	c.AppendContextualRule(rule...)
	return c
}
