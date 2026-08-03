package casing

import "unicode"

func hasPluralInitialismSuffix(runes []rune, index int) bool {
	if index < 0 || index+1 >= len(runes) {
		return false
	}
	return unicode.IsUpper(runes[index]) && unicode.IsLower(runes[index+1]) && index+1 == len(runes)-1 && runes[index+1] == 's'
}

func hasLowerPrefixAcronymBoundary(runes []rune, index int) bool {
	if index != 1 || len(runes) < 3 {
		return false
	}
	if !unicode.IsLower(runes[index-1]) || !unicode.IsUpper(runes[index]) || index+1 >= len(runes) || !unicode.IsUpper(runes[index+1]) {
		return false
	}
	for i := index; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) || unicode.IsDigit(runes[i]) {
			continue
		}
		return i == len(runes)-1 && runes[i] == 's'
	}
	return true
}

func splitPluralInitialismSuffix(remainder string) ([]string, bool) {
	if remainder == "s" {
		return []string{"s"}, true
	}
	return nil, false
}
