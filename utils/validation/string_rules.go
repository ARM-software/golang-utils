package validation

// string_rules.go contains string- and byte-oriented validation helpers that
// are not part of the schema-vocabulary layer but are still useful as reusable
// ozzo-style rules.

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Prefix validates that a string or byte slice starts with prefix.
func Prefix(prefix string) validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return strings.HasPrefix(value, prefix)
	}, fmt.Sprintf("must start with %q", prefix))
}

// Suffix validates that a string or byte slice ends with suffix.
func Suffix(suffix string) validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return strings.HasSuffix(value, suffix)
	}, fmt.Sprintf("must end with %q", suffix))
}

// ContainsString validates that a string or byte slice contains substring.
func ContainsString(substring string) validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return strings.Contains(value, substring)
	}, fmt.Sprintf("must contain %q", substring))
}

// Like validates that a string or byte slice matches re.
func Like(re *regexp.Regexp) validation.Rule {
	return validation.Match(re)
}

// NotContains validates that a string or byte slice does not contain substring.
func NotContains(substring string) validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return !strings.Contains(value, substring)
	}, fmt.Sprintf("must not contain %q", substring))
}

// NotContainsWhitespaces validates that a string or byte slice contains no whitespace characters.
func NotContainsWhitespaces() validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return strings.IndexFunc(value, unicode.IsSpace) < 0
	}, "must not contain whitespaces")
}

// MinOccurs validates that substring occurs at least min times in a string or
// byte slice.
func MinOccurs(substring string, min int) validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return strings.Count(value, substring) >= min
	}, fmt.Sprintf("must contain %q at least %d times", substring, min))
}

// MaxOccurs validates that substring occurs at most max times in a string or
// byte slice.
func MaxOccurs(substring string, max int) validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return strings.Count(value, substring) <= max
	}, fmt.Sprintf("must contain %q at most %d times", substring, max))
}

// OccursExactly validates that substring occurs exactly count times in a string
// or byte slice.
func OccursExactly(substring string, count int) validation.Rule {
	return validation.NewStringRule(func(value string) bool {
		return strings.Count(value, substring) == count
	}, fmt.Sprintf("must contain %q exactly %d times", substring, count))
}
