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

func SliceTake[T any](slice []T, n int) (result []T) {
	min := Min(len(slice), n)
	for i := 0; i < min; i++ {
		result = append(result, slice[i])
	}
	return
}

func SliceSkip[T any](slice []T, n int) (result []T) {
	if n > 0 && n < len(slice) {
		result = slice[n:]
	} else {
		result = slice
	}
	return
}
