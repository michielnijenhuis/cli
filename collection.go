package cli

type Collection[T any] []T

func Collect[T any](data []T) Collection[T] {
	return data
}

func (c Collection[T]) Len() int {
	return len(c)
}

func Map[T any, U any](collection Collection[T], fn func(T) U) Collection[U] {
	s := make([]U, collection.Len())

	return s
}

func (c Collection[T]) Map(fn func(T) T) Collection[T] {
	s := make([]T, c.Len())
	for _, v := range c {
		s = append(s, fn(v))
	}
	return Collect(s)
}

func (c Collection[T]) MapInPlace(fn func(T) T) Collection[T] {
	for i, v := range c {
		c[i] = fn(v)
	}
	return c
}

func (c Collection[T]) Filter(predicate func(T) bool) Collection[T] {
	f := make([]T, 0, c.Len())
	for _, v := range c {
		if predicate(v) {
			f = append(f, v)
		}
	}
	return Collect(f)
}

func noop[T any](v T) T {
	return v
}

func (c Collection[T]) FilterInPlace(predicate func(T) bool) Collection[T] {
	return c.FilterMapInPlace(predicate, noop)
}

func (c Collection[T]) FilterMapInPlace(predicate func(T) bool, mapper func(T) T) Collection[T] {
	var ptr int
	for _, v := range c {
		if predicate(v) {
			c[ptr] = mapper(v)
			ptr++
		}
	}
	c = c[:ptr]
	return c
}

func (c *Collection[T]) PushBack(v ...T) {
	*c = append(*c, v...)
}

func (c *Collection[T]) PushFront(v ...T) {
	s := make([]T, 0, len(v)+c.Len())
	s = append(s, v...)
	s = append(s, *c...)
	*c = s
}

func (c *Collection[T]) PopBack() T {
	last := (*c)[len(*c)-1]
	*c = (*c)[:len(*c)-1]
	return last
}

func (c *Collection[T]) PopFront() T {
	if len(*c) == 0 {
		var empty T
		return empty
	}
	first := (*c)[0]
	*c = (*c)[1:]
	return first
}

func (c Collection[T]) Has(value T, cmp func(T, T) bool) bool {
	for _, v := range c {
		if cmp(v, value) {
			return true
		}
	}
	return false
}

func HasString(c Collection[string], v string) bool {
	return c.Has(v, func(a string, b string) bool {
		return a == b
	})
}

func HasInt(c Collection[int], v int) bool {
	return c.Has(v, func(a int, b int) bool {
		return a == b
	})
}

func (c Collection[T]) IndexOf(value T, cmp func(T, T) bool) int {
	for i, v := range c {
		if cmp(v, value) {
			return i
		}
	}
	return -1
}

func (c Collection[T]) Find(fn func(T) bool) (T, bool) {
	for _, v := range c {
		if fn(v) {
			return v, true
		}
	}
	var empty T
	return empty, false
}

func (c Collection[T]) Splice(start int, end int) Collection[T] {
	if start < 0 {
		panic("start must be at least 0")
	}

	if end < start {
		panic("end must be at least start")
	}

	if end >= len(c) {
		end = len(c) - 1
	}

	s := make([]T, end-start+1)
	for i := start; i <= end; i++ {
		s = append(s, c[i])
	}

	return Collect(s)
}

func (c Collection[T]) Reverse() Collection[T] {
	r := make([]T, 0, len(c))
	for i := len(c) - 1; i >= 0; i-- {
		r = append(r, c[i])
	}
	return Collect(r)
}

func (c Collection[T]) ReverseInPlace() Collection[T] {
	j := len(c) - 1
	for i := 0; i < j; i++ {
		c[i], c[j] = c[j], c[i]
		j--
	}
	return c
}

func Reduce[T any, U any](c Collection[T], fn func(acc U, cur T, i int) U, initial U) U {
	for i, v := range c {
		initial = fn(initial, v, i)
	}

	return initial
}

func (c Collection[T]) Every(fn func(T) bool) bool {
	for _, v := range c {
		if !fn(v) {
			return false
		}
	}
	return true
}

func (c Collection[T]) Some(fn func(T) bool) bool {
	for _, v := range c {
		if fn(v) {
			return true
		}
	}
	return false
}
