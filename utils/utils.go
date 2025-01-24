package utils

// SliceUnique remove duplicate elements in slice.
// Play: https://go.dev/play/p/AXw0R3ZTE6a
func SliceUnique[T comparable](slice []T) []T {
	result := make([]T, 0, len(slice))
	seen := make(map[T]struct{}, len(slice))

	for i := range slice {
		if _, ok := seen[slice[i]]; ok {
			continue
		}

		seen[slice[i]] = struct{}{}

		result = append(result, slice[i])
	}

	return result
}
