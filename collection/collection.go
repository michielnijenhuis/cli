package collection

type Collection[T any] struct {
	data []T
}

func Collect[T any](data []T) *Collection[T] {
	return &Collection[T]{data}
}

func (c *Collection[T]) Len() int {
	return len(c.data)
}

func Map[T any, U any](collection Collection[T], fn func(T) U) Collection[U] {
	s := make([]U, collection.Len())

	return Collection[U]{data: s}
}

func (c *Collection[T]) Map(fn func(T) T) *Collection[T] {
	s := make([]T, c.Len())
	for _, v := range c.data {
		s = append(s, fn(v))
	}
	return Collect(s)
}

func (c *Collection[T]) MapInPlace(fn func(T) T) *Collection[T] {
	for i, v := range c.data {
		c.data[i] = fn(v)
	}
	return c
}

func (c *Collection[T]) Filter(predicate func(T) bool) *Collection[T] {
	f := make([]T, 0, c.Len())
	for _, v := range c.Slice() {
		if predicate(v) {
			f = append(f, v)
		}
	}
	return Collect(f)
}

func (c *Collection[T]) FilterInPlace(predicate func(T) bool) *Collection[T] {
	var ptr int
	for _, v := range c.data {
		if predicate(v) {
			c.data[ptr] = v
			ptr++
		}
	}
	c.data = c.data[:ptr]
	return c
}

func (c *Collection[T]) FilterMapInPlace(predicate func(T) bool, mapper func(T) T) *Collection[T] {
	var ptr int
	for _, v := range c.data {
		if predicate(v) {
			c.data[ptr] = mapper(v)
			ptr++
		}
	}
	c.data = c.data[:ptr]
	return c
}

func (c *Collection[T]) PushBack(v ...T) {
	c.data = append(c.data, v...)
}

func (c *Collection[T]) PushFront(v ...T) {
	s := make([]T, 0, len(v)+c.Len())
	s = append(s, v...)
	s = append(s, c.data...)
	c.data = s
}

func (c *Collection[T]) PopBack() T {
	last := c.data[len(c.data)-1]
	c.data = c.data[:len(c.data)-1]
	return last
}

func (c *Collection[T]) PopFront() T {
	if len(c.data) == 0 {
		var empty T
		return empty
	}
	first := c.data[0]
	c.data = c.data[1:]
	return first
}

func (c *Collection[T]) Has(value T, cmp func(T, T) bool) bool {
	for _, v := range c.data {
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

func (c *Collection[T]) IndexOf(value T, cmp func(T, T) bool) int {
	for i, v := range c.data {
		if cmp(v, value) {
			return i
		}
	}
	return -1
}

func (c *Collection[T]) Find(fn func(T) bool) T {
	for _, v := range c.data {
		if fn(v) {
			return v
		}
	}
	var empty T
	return empty
}

func (c *Collection[T]) Slice() []T {
	return c.data
}

func (c *Collection[T]) Splice(start int, end int) *Collection[T] {
	if start < 0 {
		panic("start must be at least 0")
	}

	if end < start {
		panic("end must be at least start")
	}

	if end >= len(c.data) {
		end = len(c.data) - 1
	}

	s := make([]T, end-start+1)
	for i := start; i <= end; i++ {
		s = append(s, c.data[i])
	}

	return Collect(s)
}

func (c *Collection[T]) Reverse() *Collection[T] {
	r := make([]T, 0, len(c.data))
	for i := len(c.data) - 1; i >= 0; i-- {
		r = append(r, c.data[i])
	}
	return Collect(r)
}
