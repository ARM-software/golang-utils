package casing

import (
	"strings"
	"unicode"

	"github.com/ARM-software/golang-utils/utils/collection"
)

// This codebase uses a leading `I` to name interfaces, so identifiers such as
// `IHTTP` should be recognised as an interface prefix plus acronym rather than
// as a single undifferentiated all-caps token.

func hasInterfacePrefixAcronymBoundary(runes []rune, index int) bool {
	if index != 1 || len(runes) < 3 || runes[0] != 'I' {
		return false
	}
	return unicode.IsUpper(runes[index]) && index+1 < len(runes) && unicode.IsUpper(runes[index+1])
}

func isInterfacePrefixedAcronym(value string) bool {
	runes := []rune(value)
	if len(runes) < 3 || runes[0] != 'I' {
		return false
	}
	return collection.AllFunc(runes[1:], func(r rune) bool {
		return unicode.IsUpper(r) || unicode.IsDigit(r)
	})
}

func normaliseInterfacePrefixedAcronym(value string, replacer *Replacer, lowerPrefix bool) (string, bool) {
	if replacer == nil || len(value) < 2 {
		return "", false
	}
	runes := []rune(value)
	if runes[0] != 'I' && runes[0] != 'i' {
		return "", false
	}
	parts, ok := replacer.compoundReplacementParts(strings.ToLower(string(runes[1:])))
	if !ok || len(parts) == 0 {
		return "", false
	}
	prefix := "I"
	if lowerPrefix {
		prefix = "i"
	}
	return prefix + strings.Join(parts, ""), true
}
