package util

func SliceFrom[T any](slice []T, idx int) (ret []T, ok bool) {
	if idx < 0 || idx >= len(slice) {
		return []T{}, false
	}
	return slice[idx:], true
}

func SliceTo[T any](slice []T, idx int) (ret []T, ok bool) {
	if idx < 0 || idx >= len(slice) {
		return []T{}, false
	}
	return slice[:idx], true
}

func SliceStrFrom(slice string, idx int) (ret string, ok bool) {
	if idx < 0 || idx >= len(slice) {
		return "", false
	}
	return slice[idx:], true
}

func SliceStrTo(slice string, idx int) (ret string, ok bool) {
	if idx < 0 || idx >= len(slice) {
		return "", false
	}
	return slice[:idx], true
}

// Contains checks if a string slice contains a specific string
func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
