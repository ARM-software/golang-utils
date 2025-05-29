package maps

import (
	"strings"
)

const separator = "."

// Map is a wrapper around maps[string]string that provides some helpers
// above it that assume the maps is in the format that flatmap expects
// (the result of Flatten).
//
// All modifying functions such as Delete are done in-place unless
// otherwise noted.
type Map map[string]string

// Contains returns true if the maps contains the given key.
func (m Map) Contains(key string) bool {
	for _, k := range m.Keys() {
		if k == key {
			return true
		}
	}

	return false
}

func (m Map) AsMap() map[string]string {
	return m
}

// Delete deletes a key out of the maps with the given prefix.
func (m Map) Delete(prefix string) {
	for k, _ := range m {
		match := k == prefix
		if !match {
			if !strings.HasPrefix(k, prefix) {
				continue
			}

			if k[len(prefix):len(prefix)+1] != separator {
				continue
			}
		}

		delete(m, k)
	}
}

// Keys returns all the top-level keys in this maps
func (m Map) Keys() []string {
	ks := make(map[string]struct{})
	for k, _ := range m {
		idx := strings.Index(k, separator)
		if idx == -1 {
			idx = len(k)
		}

		ks[k[:idx]] = struct{}{}
	}

	result := make([]string, 0, len(ks))
	for k, _ := range ks {
		result = append(result, k)
	}

	return result
}

// Merge merges the contents of the other Map into this one.
//
// This merge is smarter than a simple maps iteration because it
// will fully replace arrays and other complex structures that
// are present in this maps with the other maps's. For example, if
// this maps has a 3 element "foo" list, and m2 has a 2 element "foo"
// list, then the result will be that m has a 2 element "foo"
// list.
func (m Map) Merge(m2 Map) {
	for _, prefix := range m2.Keys() {
		m.Delete(prefix)

		for k, v := range m2 {
			if strings.HasPrefix(k, prefix) {
				m[k] = v
			}
		}
	}
}
