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
