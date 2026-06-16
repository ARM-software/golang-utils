package collection

import (
	"iter"
	"slices"
)

// IndexBy builds a map from derived key to value.
//
// If multiple elements produce the same key, the last one wins.
func IndexBy[S ~[]E, E any, K comparable](slice S, keyFunc KeyFunc[E, K]) map[K]E {
	return IndexBySequence(slices.Values(slice), keyFunc)
}

// IndexByRef behaves like IndexBy but accepts a reference-based key function.
func IndexByRef[S ~[]E, E any, K comparable](slice S, keyFunc KeyRefFunc[E, K]) map[K]E {
	return IndexBy(slice, toKeyFunc(keyFunc))
}

// IndexBySequence builds a map from derived key to value from a sequence.
//
// If multiple elements produce the same key, the last one wins.
func IndexBySequence[E any, K comparable](sequence iter.Seq[E], keyFunc KeyFunc[E, K]) map[K]E {
	return ReducesSequence(sequence, map[K]E{}, func(acc map[K]E, element E) map[K]E {
		acc[keyFunc(element)] = element
		return acc
	})
}

// IndexByRefSequence behaves like IndexBySequence but accepts a reference-based key function.
func IndexByRefSequence[E any, K comparable](sequence iter.Seq[E], keyFunc KeyRefFunc[E, K]) map[K]E {
	return IndexBySequence(sequence, toKeyFunc(keyFunc))
}
