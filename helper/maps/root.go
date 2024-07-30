package maps

func Keys[T any](m map[string]T) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
