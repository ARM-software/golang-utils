package casing

import (
	"strings"
	"unicode"

	"github.com/sttk/stringcase"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// ToCamelCase converts value to camelCase and optionally applies a replacer to the resulting identifier. Only the first replacer is used.
func ToCamelCase(value string, replacers ...*Replacer) string {
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		if normalised, ok := normaliseInterfacePrefixedAcronym(value, replacer, true); ok {
			return normalised
		}
		if isIdentifierWithoutSeparators(value) {
			return lowerFirstWord(replaceIdentifierWords(value, replacer))
		}
		result := stringcase.CamelCase(prepareCaseInput(value))
		return replacer.Replace(result)
	}
	return stringcase.CamelCase(prepareCaseInput(value))
}

// ToPascalCase converts value to PascalCase and optionally applies a replacer to the resulting identifier. Only the first replacer is used.
func ToPascalCase(value string, replacers ...*Replacer) string {
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		if normalised, ok := normaliseInterfacePrefixedAcronym(value, replacer, false); ok {
			return normalised
		}
		if isIdentifierWithoutSeparators(value) {
			return replaceIdentifierWords(value, replacer)
		}
		result := stringcase.PascalCase(prepareCaseInput(value))
		return replacer.Replace(result)
	}
	return stringcase.PascalCase(prepareCaseInput(value))
}

// ToSnakeCase converts value to snake_case and optionally applies a replacer to the identifier before the final case conversion. Only the first replacer is used.
func ToSnakeCase(value string, replacers ...*Replacer) string {
	result := value
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		if normalised, ok := normaliseInterfacePrefixedAcronym(value, replacer, true); ok {
			return strings.ToLower(normalised)
		}
		if isIdentifierWithoutSeparators(value) {
			result = replaceIdentifierWords(value, replacer)
		} else {
			result = replacer.Replace(stringcase.PascalCase(prepareCaseInput(value)))
		}
		if isInterfacePrefixedAcronym(result) {
			return strings.ToLower(result)
		}
		return strings.ToLower(strings.Join(splitCamelWords(result), "_"))
	}
	return stringcase.SnakeCase(result)
}

// ToKebabCase converts value to kebab-case and optionally applies a replacer to the identifier before the final case conversion. Only the first replacer is used.
func ToKebabCase(value string, replacers ...*Replacer) string {
	result := value
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		if normalised, ok := normaliseInterfacePrefixedAcronym(value, replacer, true); ok {
			return strings.ToLower(normalised)
		}
		if isIdentifierWithoutSeparators(value) {
			result = replaceIdentifierWords(value, replacer)
		} else {
			result = replacer.Replace(stringcase.PascalCase(prepareCaseInput(value)))
		}
		if isInterfacePrefixedAcronym(result) {
			return strings.ToLower(result)
		}
		return strings.ToLower(strings.Join(splitCamelWords(result), "-"))
	}
	return stringcase.KebabCase(result)
}

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

func replaceIdentifierWords(value string, replacer *Replacer) string {
	if parts, ok := replacer.compoundReplacementParts(strings.ToLower(value)); ok && len(parts) > 0 {
		return strings.Join(parts, "")
	}
	parts := splitCamelWords(value)
	if len(parts) == 0 {
		return value
	}
	var builder strings.Builder
	collection.ForEach(parts, func(part string) {
		builder.WriteString(replacer.transformWord(part, builder.Len() == 0, false))
	})
	return builder.String()
}

func lowerFirstWord(value string) string {
	parts := splitCamelWords(value)
	if len(parts) == 0 {
		return value
	}
	parts[0] = strings.ToLower(parts[0])
	return strings.Join(parts, "")
}
