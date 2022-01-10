package stream

type Stream[T any] interface {
	Filter(predicate func(element T) bool) Stream[T]
}

type emptyStream[T any] struct{}

func (es *emptyStream[T]) Filter(func(T) bool) Stream[T] {
	return es
}

type singletonStream[T any] struct {
	value T
}

func (s *singletonStream[T]) Filter(predicate func(element T) bool) Stream[T] {
	if predicate(s.value) {
		return s
	}
	return &emptyStream[T]{}
}
