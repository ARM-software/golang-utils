package casing

import (
	"strings"
	"unicode"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

func prepareCaseInput(value string) string {
	if !isIdentifierWithoutSeparators(value) || !hasUppercase(value) {
		return value
	}
	parts := splitCamelWords(value)
	if len(parts) <= 1 {
		return value
	}
	return strings.Join(parts, "_")
}

func splitCamelWords(value string) []string {
	if reflection.IsEmpty(value) {
		return nil
	}
	runes := []rune(value)
	parts := make([]string, 0)
	start := 0
	for i := 1; i < len(runes); i++ {
		prev := runes[i-1]
		curr := runes[i]
		nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
		acronymPluralSuffix := hasPluralInitialismSuffix(runes, i)
		if hasLowerPrefixAcronymBoundary(runes, i) {
			continue
		}
		if hasInterfacePrefixAcronymBoundary(runes, i) {
			parts = append(parts, string(runes[start:i]))
			start = i
			continue
		}
		if unicode.IsLower(prev) && unicode.IsUpper(curr) || unicode.IsDigit(prev) && unicode.IsUpper(curr) || unicode.IsUpper(prev) && unicode.IsUpper(curr) && nextIsLower && !acronymPluralSuffix {
			parts = append(parts, string(runes[start:i]))
			start = i
		}
	}
	parts = append(parts, string(runes[start:]))
	return parts
}

func formSnakeCasedWords(parts []string) string {
	return strings.ToLower(strings.Join(parts, "_"))
}

func formKebabCasedWords(parts []string) string {
	return strings.ToLower(strings.Join(parts, "-"))
}

func isIdentifierWithoutSeparators(value string) bool {
	if reflection.IsEmpty(value) {
		return false
	}
	return collection.AllFunc([]rune(value), func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	})
}

func hasUppercase(value string) bool {
	return collection.AnyFunc([]rune(value), unicode.IsUpper)
}

func startsWithLowercase(value string) bool {
	first, ok := collection.First([]rune(value))
	if !ok {
		return false
	}
	return unicode.IsLower(first)
}

func replaceIdentifierWords(value string, replacer *Replacer, firstWordLowercase bool) string {
	if parts, ok := replacer.splitAdjacentTokenReplacementParts(strings.ToLower(value)); ok && len(parts) > 0 {
		joined := strings.Join(parts, "")
		if firstWordLowercase {
			parts[0] = strings.ToLower(parts[0])
			return strings.Join(parts, "")
		}
		return upperFirstWord(joined)
	}
	if parts, ok := splitLeadingLetterCompound(value, replacer); ok {
		joined := strings.Join(parts, "")
		if firstWordLowercase {
			return lowerFirstWord(joined)
		}
		return upperFirstWord(joined)
	}
	parts := splitCamelWords(value)
	if reflection.IsEmpty(parts) {
		return value
	}
	var builder strings.Builder
	collection.ForEach(parts, func(part string) {
		builder.WriteString(replacer.transformWord(part, builder.Len() == 0, firstWordLowercase))
	})
	return builder.String()
}

func lowerFirstWord(value string) string {
	runes := []rune(value)
	first, ok := collection.First(runes)
	if !ok {
		return value
	}
	runes[0] = unicode.ToLower(first)
	return string(runes)
}

func upperFirstWord(value string) string {
	runes := []rune(value)
	first, ok := collection.First(runes)
	if !ok {
		return value
	}
	runes[0] = unicode.ToUpper(first)
	return string(runes)
}

func isUpperInitialismOrPlural(value string) bool {
	runes := []rune(value)
	last, ok := collection.Last(runes)
	if !ok {
		return false
	}
	body := runes
	if last == 's' {
		body = runes[:len(runes)-1]
	}
	return collection.AllFunc(body, func(r rune) bool {
		return unicode.IsUpper(r) || unicode.IsDigit(r)
	})
}

func splitLeadingLetterCompound(value string, replacer *Replacer) ([]string, bool) {
	runes := []rune(value)
	if replacer == nil || len(runes) < 2 || !unicode.IsLetter(runes[0]) || !unicode.IsUpper(runes[1]) {
		return nil, false
	}
	leadingParts := splitCamelWords(value)
	if len(leadingParts) > 0 {
		if parts, ok := replacer.splitAdjacentTokenReplacementParts(strings.ToLower(leadingParts[0])); ok && len(parts) > 0 {
			return nil, false
		}
	}
	tailWords := splitCamelWords(string(runes[1:]))
	if len(tailWords) < 2 {
		return nil, false
	}
	if len([]rune(tailWords[0])) < 2 {
		return nil, false
	}
	firstTail := replacer.transformWord(tailWords[0], true, false)
	if !isUpperInitialismOrPlural(firstTail) {
		return nil, false
	}
	parts := append([]string{string(runes[0]), firstTail}, collection.Map(tailWords[1:], func(word string) string {
		return replacer.transformWord(word, false, false)
	})...)
	return parts, true
}

func splitNumericInitialismSuffix(remainder string) ([]string, bool) {
	if reflection.IsEmpty(remainder) {
		return nil, false
	}
	if !collection.AllFunc([]rune(remainder), unicode.IsDigit) {
		return nil, false
	}
	return []string{remainder}, true
}
