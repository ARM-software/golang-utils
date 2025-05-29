package maps

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func Expand(value map[string]string) (expandedMap any, err error) {
	if len(value) == 0 {
		return
	}
	m := make(map[string]any, len(value))
	for k := range value {
		key, _, found := strings.Cut(k, separator)
		if found {
			subMap, subErr := ExpandPrefixed(value, key)
			if subErr != nil {
				err = subErr
				return
			}
			m[key] = subMap
		} else {
			m[k] = value[k]
		}
	}
	expandedMap = m
	return
}

// ExpandPrefixed takes a maps and a prefix and expands that value into
// a more complex structure. This is the reverse of the Flatten operation.
func ExpandPrefixed(m map[string]string, key string) (result any, err error) {
	// If the key is exactly a key in the maps, just return it
	if v, ok := m[key]; ok {
		if v == "true" {
			result = true
			return
		} else if v == "false" {
			result = false
			return
		}

		result = v
		return
	}

	// Check if the key is an array, and if so, expand the array
	arrayKey := fmt.Sprintf("%v%v%d", key, separator, 0)
	if _, ok := m[arrayKey]; ok {
		result, err = expandArray(m, key)
		return
	}
	arrayKey = fmt.Sprintf("%v%v", arrayKey, separator)
	for k := range m {
		if strings.HasPrefix(k, arrayKey) {
			result, err = expandArray(m, key)
			return
		}
	}

	// Check if this is a prefix in the maps
	prefix := key + separator
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			result, err = expandMap(m, prefix)
			return
		}
	}

	result = nil
	return
}

func expandArray(m map[string]string, prefix string) (result []any, err error) {
	keySet := map[int]bool{}
	for k := range m {
		if !strings.HasPrefix(k, prefix+separator) {
			continue
		}

		key := k[len(prefix)+1:]
		idx := strings.Index(key, separator)
		if idx != -1 {
			key = key[:idx]
		}

		k, subErr := strconv.Atoi(key)
		if subErr != nil {
			err = subErr
			return
		}
		keySet[k] = true
	}

	var keysList []int
	for key := range keySet {
		keysList = append(keysList, key)
	}
	sort.Ints(keysList)

	r := make([]any, len(keysList))
	for i, key := range keysList {
		keyString := strconv.Itoa(key)
		item, subErr := ExpandPrefixed(m, fmt.Sprintf("%s.%s", prefix, keyString))
		r[i] = item
		if subErr != nil {
			err = subErr
			return
		}
	}
	result = r
	return
}

func expandMap(m map[string]string, prefix string) (r map[string]any, err error) {
	result := make(map[string]any)
	for k := range m {
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		key := k[len(prefix):]
		idx := strings.Index(key, separator)
		if idx != -1 {
			key = key[:idx]
		}
		if _, ok := result[key]; ok {
			continue
		}

		item, subErr := ExpandPrefixed(m, k[:len(prefix)+len(key)])
		if subErr != nil {
			err = subErr
			return
		}
		result[key] = item
	}

	r = result
	return
}
