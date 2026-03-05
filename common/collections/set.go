package collections

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](arr []T) *Set[T] {
	m := map[T]struct{}{}
	for _, item := range arr {
		m[item] = struct{}{}
	}

	return &Set[T]{
		m: m,
	}
}

func (s *Set[T]) Has(item T) bool {
	_, ok := s.m[item]
	return ok
}

func (s *Set[T]) Add(item T) {
	s.m[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	delete(s.m, item)
}

func (s *Set[T]) Len() int {
	return len(s.m)
}

func (s *Set[T]) GetValues() []T {
	values := make([]T, 0, len(s.m))
	for key := range s.m {
		values = append(values, key)
	}

	return values
}
