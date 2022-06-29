package maputil

// Clone a generic map. This is NOT a deep clone.
func Clone[K comparable, V any, T map[K]V](m T) T {
	result := make(T, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
