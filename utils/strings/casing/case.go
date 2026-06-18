package casing

import (
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/sttk/stringcase"
)

// ToCamelCase converts value to camelCase and optionally applies a replacer to the resulting identifier. Only the first replacer is used.
func ToCamelCase(value string, replacers ...*Replacer) string {
	result := stringcase.CamelCase(value)
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		return replacer.Replace(result)
	}
	return result
}

// ToPascalCase converts value to PascalCase and optionally applies a replacer to the resulting identifier. Only the first replacer is used.
func ToPascalCase(value string, replacers ...*Replacer) string {
	result := stringcase.PascalCase(value)
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		return replacer.Replace(result)
	}
	return result
}

// ToSnakeCase converts value to snake_case and optionally applies a replacer to the identifier before the final case conversion. Only the first replacer is used.
func ToSnakeCase(value string, replacers ...*Replacer) string {
	result := value
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		result = replacer.Replace(stringcase.PascalCase(value))
	}
	return stringcase.SnakeCase(result)
}

// ToKebabCase converts value to kebab-case and optionally applies a replacer to the identifier before the final case conversion. Only the first replacer is used.
func ToKebabCase(value string, replacers ...*Replacer) string {
	result := value
	if replacer, ok := collection.First(replacers); ok && replacer != nil {
		result = replacer.Replace(stringcase.PascalCase(value))
	}
	return stringcase.KebabCase(result)
}
