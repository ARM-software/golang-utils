package safecast

func greaterThanUpperBoundary[C1 IConvertable, C2 IConvertable](value C1, upperBoundary C2) (greater bool) {
	if value <= 0 {
		return
	}

	switch f := any(value).(type) {
	case float64:
		greater = f >= float64(upperBoundary)
	case float32:
		greater = float64(f) >= float64(upperBoundary)
	default:
		// for all other integer types, it fits in an uint64 without overflow as we know value is positive.
		greater = uint64(value) > uint64(upperBoundary)
	}

	return
}

func lessThanLowerBoundary[T IConvertable, T2 IConvertable](value T, boundary T2) (lower bool) {
	if value >= 0 {
		return
	}

	switch f := any(value).(type) {
	case float64:
		lower = f <= float64(boundary)
	case float32:
		lower = float64(f) <= float64(boundary)
	default:
		lower = int64(value) < int64(boundary)
	}
	return
}
