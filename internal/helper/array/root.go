package array

import "sort"

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

func Remove[T comparable](s []T, v T) []T {
	i := IndexOf(s, v)
	if i == -1 {
		return s
	}

	return append(s[:i], s[i+1:]...)
}

func IndexOfFunc[T any](s []T, v T, cmp func(T, T) bool) int {
	for i, x := range s {
		if cmp(x, v) {
			return i
		}
	}

	return -1
}

func Keys[T comparable, U any](m map[T]U) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func SortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
