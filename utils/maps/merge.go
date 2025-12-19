package maps

import "maps"

// Merge merges multiple maps into a new map.
// Later maps override earlier ones on key conflicts.
// Nil maps are ignored.
func Merge[K comparable, T any](m ...map[K]T) map[K]T {
	switch len(m) {
	case 0:
		return map[K]T{}
	case 1:
		merge := m[0]
		if merge == nil {
			merge = map[K]T{}
		}
		return merge
	}

	dest := map[K]T{}
	for i := range m {
		src := m[i]
		if src == nil {
			continue
		}
		maps.Copy(dest, src)
	}
	return dest
}
