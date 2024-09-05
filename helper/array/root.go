package array

func IncludesString(s []string, v string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == v {
			return true
		}
	}

	return false
}

func IndexOf[T comparable](s []T, v T) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}

	return -1
}

func IndexOfFunc[T any](s []T, v T, cmp func(T, T) bool) int {
	for i, x := range s {
		if cmp(x, v) {
			return i
		}
	}

	return -1
}
