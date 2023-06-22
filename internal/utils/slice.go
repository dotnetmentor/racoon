package utils

func StringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func SliceContains[T any](slice []T, match func(i T) bool) bool {
	for _, i := range slice {
		if match(i) {
			return true
		}
	}
	return false
}

func SliceDelete[T any](slice []T, match func(i T) bool) (result []T) {
	for _, i := range slice {
		if !match(i) {
			result = append(result, i)
		}
	}
	return
}
