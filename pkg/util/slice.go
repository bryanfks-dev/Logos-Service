package util

func CoalesceSlice[T any](primary, fallback []T) []T {
	if len(primary) > 0 {
		return primary
	}

	return fallback
}

func NonNilSlice[T any](items []T) []T {
	if len(items) == 0 {
		return make([]T, 0)
	}

	return items
}
