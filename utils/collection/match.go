package collection

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
