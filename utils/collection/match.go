package collection

import (
	"iter"
	"slices"
)

//
// Match utilities
//

// match applies each match function to e and returns a Conditions containing the outcomes.
func match[E any](e E, matches []FilterFunc[E]) *Conditions {
	conditions := NewConditions(len(matches))
	for i := range matches {
		conditions.Add(matches[i](e))
	}
	return &conditions
}

// Match returns true if any of the provided match predicates return true for e.
func Match[E any](e E, matches ...FilterFunc[E]) bool {
	return match[E](e, matches).Any()
}

// MatchAll returns true only if all the provided match predicates return true for e.
func MatchAll[E any](e E, matches ...FilterFunc[E]) bool {
	return match[E](e, matches).All()
}

// InSequence reports whether any element in s matches v using any of the
// provided match functions. Match functions that return an error are treated
// as non-matches.
func InSequence[E any](s iter.Seq[E], v E, m ...MatchFunc[E]) bool {
	matching := Map(m, func(mf MatchFunc[E]) Predicate[E] {
		return matchToPredicateFunc(v, mf)
	})
	return matchInSeq(s, matching)
}

// InSequenceRef behaves like InSequence but uses reference-based match
// functions and a reference value.
func InSequenceRef[E any](s iter.Seq[E], v *E, m ...MatchRefFunc[E]) bool {
	matching := Map(m, func(mf MatchRefFunc[E]) Predicate[E] {
		return toPredicateFunc(matchToPredicateRefFunc(v, mf))
	})
	return matchInSeq(s, matching)
}

// In reports whether any element in s matches v using any of the provided
// match functions. Match functions that return an error are treated as
// non-matches.
func In[Slice ~[]E, E any](s Slice, v E, m ...MatchFunc[E]) bool {
	return InSequence(slices.Values(s), v, m...)
}

// InRef behaves like In but uses reference-based match functions and a
// reference value.
func InRef[Slice ~[]E, E any](s Slice, v *E, m ...MatchRefFunc[E]) bool {
	return InSequenceRef(slices.Values(s), v, m...)
}

func matchInSeq[E any](s iter.Seq[E], matching []Predicate[E]) bool {
	matchingF := func(e E) bool {
		return Match(e, matching...)
	}
	_, found := FindInSequence(s, matchingF)
	return found
}
