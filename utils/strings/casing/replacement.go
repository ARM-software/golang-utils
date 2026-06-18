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

// Rule defines a token replacement rule for identifier-like strings.
type Rule struct {
	Token       string
	Replacement string
	Exceptions  []string
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
	compiled := make(map[string]compiledRule, len(rules))
	err := collection.Each(slices.Values(rules), func(rule Rule) error {
		if err := rule.Validate(); err != nil {
			return err
		}

		exceptions := mapset.NewSet[string]()
		normalised := collection.Map(rule.Exceptions, func(exception string) string {
			return strings.ToLower(strings.TrimSpace(exception))
		})
		_ = exceptions.Append(normalised...)

		compiled[strings.ToLower(strings.TrimSpace(rule.Token))] = compiledRule{
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
