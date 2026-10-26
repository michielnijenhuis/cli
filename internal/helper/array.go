package helper

func Shift[T any](tokens *[]T) T {
	if len(*tokens) == 0 {
		var empty T
		return empty
	}

	first := (*tokens)[0]
	slice := make([]T, 0, len(*tokens)-1)
	slice = append(slice, (*tokens)[1:]...)
	*tokens = slice

	return first
}

func Unshift[T any](tokens *[]T, value T) {
	if tokens == nil {
		return
	}

	slice := make([]T, 0, len(*tokens)+1)
	slice = append(slice, value)
	slice = append(slice, (*tokens)...)
	*tokens = slice
}

func Grow[T any](s []T, l int) []T {
	if s == nil {
		return make([]T, l)
	}

	if len(s) >= l {
		return s
	}

	var zero T
	for i := 0; i < l-len(s); i++ {
		s = append(s, zero)
	}

	return s
}
