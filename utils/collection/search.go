package collection

// Looks for an element in a slice. If found it will
// return its index and true; otherwise it will return -1 and false.
func Find(slice *[]string, val string) (int, bool) {
	for i, item := range *slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func Any(slice []bool) bool {
	for _, v := range slice {
		if v {
			return true
		}
	}
	return false
}

func All(slice []bool) bool {
	for _, v := range slice {
		if !v {
			return false
		}
	}
	return true
}
