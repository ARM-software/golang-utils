package casing

import (
	"strings"

	"github.com/sttk/stringcase"

	"github.com/ARM-software/golang-utils/utils/collection"
)

// ToCamelCase converts value to camelCase and optionally applies a replacer to the resulting identifier. Only the first replacer is used.
func ToCamelCase(value string, replacers ...*Replacer) string {
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		if normalised, ok := normaliseInterfacePrefixedAcronym(value, replacer, true); ok {
			return normalised
		}
		if isIdentifierWithoutSeparators(value) {
			replaced := replaceIdentifierWords(value, replacer, true)
			if strings.ToLower(value) == value && isUpperInitialismOrPlural(replaced) {
				return strings.ToLower(replaced)
			}
			return lowerFirstWord(replaced)
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
			return replaceIdentifierWords(value, replacer, false)
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
			result = replaceIdentifierWords(value, replacer, startsWithLowercase(value))
		} else {
			result = replacer.Replace(stringcase.PascalCase(prepareCaseInput(value)))
		}
		if isInterfacePrefixedAcronym(result) {
			return strings.ToLower(result)
		}
		if parts, ok := splitLeadingLetterCompound(result, replacer); ok {
			return formSnakeCasedWords(parts)
		}
		return formSnakeCasedWords(splitCamelWords(result))
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
			result = replaceIdentifierWords(value, replacer, startsWithLowercase(value))
		} else {
			result = replacer.Replace(stringcase.PascalCase(prepareCaseInput(value)))
		}
		if isInterfacePrefixedAcronym(result) {
			return strings.ToLower(result)
		}
		if parts, ok := splitLeadingLetterCompound(result, replacer); ok {
			return formKebabCasedWords(parts)
		}
		return formKebabCasedWords(splitCamelWords(result))
	}
	return stringcase.KebabCase(result)
}
