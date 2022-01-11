package stream

type Stream[T any] interface {
	Filter(predicate func(element T) bool) Stream[T]
	Reduce(accumulator func(a, b T) T) (T, bool)
}

func Filter[T any](s Stream[T], predicate func(element T) bool) Stream[T] {
	return s.Filter(predicate)
}

func Map[T, R any](s Stream[T], mapper func(element T) R) Stream[R] {
	switch s := s.(type) {
	case *EmptyStream[T]:
		return &EmptyStream[R]{}
	case *SingletonStream[T]:
		return &SingletonStream[R]{mapper(s.Value)}
	}
	return nil
}

type EmptyStream[T any] struct{}

func (es *EmptyStream[T]) Filter(func(T) bool) Stream[T] {
	return es
}

func (es *EmptyStream[T]) Reduce(func(_, _ T) T) (T, bool) {
	var zero T
	return zero, false
}

type SingletonStream[T any] struct {
	Value T
}

func (s *SingletonStream[T]) Filter(predicate func(element T) bool) Stream[T] {
	if predicate(s.Value) {
		return s
	}
	return &EmptyStream[T]{}
}

func (s *SingletonStream[T]) Reduce(func(_, _ T) T) (T, bool) {
	return s.Value, true
}
