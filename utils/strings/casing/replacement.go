package casing

import (
	"context"
	"io"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sttk/stringcase"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

var (
	normaliseRuleTokenFunc       = collection.Combine(strings.ToLower, strings.TrimSpace)
	normaliseRuleExceptionFunc   = collection.Combine(strings.ToLower, strings.TrimSpace)
	normaliseRuleReplacementFunc = strings.TrimSpace
)

// Rule defines a token replacement rule for identifier-like strings.
type Rule struct {
	Token       string
	Replacement string
	Exceptions  []string
}

// IsCompatible returns whether r and other describe the same replacement rule
// and can therefore be merged safely.
//
// Rules are considered compatible when they target the same token and use the
// same replacement after normalisation. Exception lists do not need to match.

// If other is empty, the rules are treated as incompatible and false is
// returned immediately.
func (r Rule) IsCompatible(other *Rule) bool {
	if reflection.IsEmpty(other) {
		return false
	}
	return normaliseRuleTokenFunc(r.Token) == normaliseRuleTokenFunc(other.Token) && normaliseRuleReplacementFunc(r.Replacement) == normaliseRuleReplacementFunc(other.Replacement)
}

// Merge combines compatible rules into one rule.
//
// If the rules are not compatible, r is returned unchanged.
//
// When merged, exception lists are combined and deduplicated using the same
// normalisation semantics as the replacer.

// If other is empty, r is returned unchanged.
func (r Rule) Merge(other *Rule) Rule {
	if reflection.IsEmpty(other) {
		return r
	}
	if !r.IsCompatible(other) {
		return r
	}

	merged := r
	if reflection.IsEmpty(merged.Token) {
		merged.Token = strings.TrimSpace(other.Token)
	}
	if reflection.IsEmpty(merged.Replacement) {
		merged.Replacement = normaliseRuleReplacementFunc(other.Replacement)
	}
	merged.Exceptions = mergeRuleExceptions(r.Exceptions, other.Exceptions)
	return merged
}

// Validate checks that the rule contains the required values.
func (r *Rule) Validate() (err error) {
	err = validation.ValidateStruct(r,
		validation.Field(&r.Token, validation.Required),
		validation.Field(&r.Replacement, validation.Required),
	)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid replacement rule")
	}
	return
}

type compiledRule struct {
	Replacement string
	Exceptions  mapset.Set[string]
}

// Replacer applies token replacement rules to identifier-like strings.
//
// A replacer works at identifier word boundaries rather than by performing raw
// substring replacement. In practice, this means it splits names such as
// `OpenAiApiKey`, `openAiApiKey`, or `HTTPApiToken` into logical words, looks
// up each word against the configured rules, and then rebuilds the identifier
// while preserving its overall case style.
//
// If a word matches a configured rule, the rule's replacement is used unless
// that word is listed in the rule's exceptions. If a word does not match any
// rule, or it is exempted by an exception, it is reconstructed according to the
// identifier shape being processed:
//   - all-uppercase words are preserved as-is
//   - the first word of a camelCase identifier remains lower camel case
//   - subsequent words are emitted in PascalCase
//
// This makes Replacer useful for normalising acronyms and canonical tokens such
// as `API`, `HTTP`, `ID`, or `AI` across generated names, configuration keys,
// and code identifiers without losing the surrounding naming convention.
type Replacer struct {
	rules map[string]compiledRule
}

// NewReplacer constructs a Replacer from the provided rules.
func NewReplacer(rules ...Rule) (*Replacer, error) {
	mergedRules := MergeRules(rules...)
	compiled := make(map[string]compiledRule, len(mergedRules))
	err := collection.Each(slices.Values(mergedRules), func(rule Rule) error {
		if err := rule.Validate(); err != nil {
			return err
		}

		exceptions := mapset.NewSet[string]()
		normalised := collection.Map(rule.Exceptions, normaliseRuleExceptionFunc)
		_ = exceptions.Append(normalised...)

		compiled[normaliseRuleTokenFunc(rule.Token)] = compiledRule{
			Replacement: rule.Replacement,
			Exceptions:  exceptions,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Replacer{rules: compiled}, nil
}

// MergeRules merges compatible duplicate rules and keeps the last conflicting
// rule for a given token.
//
// Compatible rules share the same token and replacement and have their
// exception lists combined. Conflicting rules do not cause an error; the later
// rule replaces the earlier one. The order of the returned rules is not
// guaranteed.
func MergeRules(rules ...Rule) []Rule {
	return collection.MergeBy(rules,
		func(rule Rule) string {
			return normaliseRuleTokenFunc(rule.Token)
		},
		func(left, right Rule) (Rule, bool) {
			if left.IsCompatible(&right) {
				return left.Merge(&right), true
			}
			return right, true
		},
	)
}

// Replace rewrites an identifier-like string according to the configured rules while preserving camelCase or PascalCase shape.
func (r *Replacer) Replace(value string) string {
	if r == nil || reflection.IsEmpty(r.rules) || reflection.IsEmpty(value) {
		return value
	}

	parts := splitCamelWords(value)
	if len(parts) == 0 {
		return value
	}
	firstWordLowercase := strings.ToLower(string(value[0])) == string(value[0])

	var builder strings.Builder
	builder.Grow(len(value))
	collection.ForEach(parts, func(part string) {
		builder.WriteString(r.transformWord(part, builder.Len() == 0, firstWordLowercase))
	})
	return builder.String()
}

// WriteString rewrites s with Replace and writes the result to w using the repository's context-aware safe I/O helpers.
func (r *Replacer) WriteString(ctx context.Context, w io.Writer, s string) (n int, err error) {
	return safeio.WriteString(ctx, w, r.Replace(s))
}

func splitCamelWords(value string) []string {
	if reflection.IsEmpty(value) {
		return nil
	}
	words := strings.Split(stringcase.SnakeCase(value), "_")
	parts := make([]string, 0, len(words))
	start := 0
	_ = collection.Each(slices.Values(words), func(word string) error {
		end := start + len(word)
		if end > len(value) {
			parts = words
			return io.EOF
		}
		parts = append(parts, value[start:end])
		start = end
		return nil
	})
	return parts
}

func mergeRuleExceptions(groups ...[]string) []string {
	merged := mapset.NewSet[string]()
	collection.ForEach(groups, func(exceptions []string) {
		merged.Append(collection.Map(exceptions, normaliseRuleExceptionFunc)...)
	})
	return merged.ToSlice()
}

// transformWord rewrites one logical word inside an identifier while trying to preserve the identifier's overall shape.
//
// The method first normalises word for case-insensitive rule lookup. If there
// is no matching rule, or if the matching rule marks the word as an exception,
// the original word is reconstructed according to the target identifier style:
//   - all-uppercase words such as `API` are preserved as-is
//   - the first word of a camelCase identifier is emitted in lower camel case
//   - every other word is emitted in PascalCase
//
// If a rule applies, its configured replacement is used instead of rebuilding
// the word, with the only adjustment being that the first word is lowercased
// when the overall identifier should remain camelCase.
func (r *Replacer) transformWord(word string, first, firstWordLowercase bool) string {
	lower := strings.ToLower(word)
	rule, found := r.rules[lower]
	if !found {
		if word == strings.ToUpper(word) {
			return word
		}
		if first && firstWordLowercase {
			return stringcase.CamelCase(lower)
		}
		return stringcase.PascalCase(lower)
	}
	if rule.Exceptions.Contains(lower) {
		if word == strings.ToUpper(word) {
			return word
		}
		if first && firstWordLowercase {
			return stringcase.CamelCase(lower)
		}
		return stringcase.PascalCase(lower)
	}
	if first && firstWordLowercase {
		return strings.ToLower(rule.Replacement)
	}
	return rule.Replacement
}
