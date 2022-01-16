package stream

type Stream[T any] interface {
	Filter(predicate func(element T) bool) Stream[T]
	Reduce(accumulator func(a, b T) T) (T, bool)
}

func Of[T any](x ...T) Stream[T] {
	return Slice(x)
}

func Slice[T any](x []T) Stream[T] {
	return nil
}

func Filter[T any](s Stream[T], predicate func(element T) bool) Stream[T] {
	return s.Filter(predicate)
}

func Map[T, R any](s Stream[T], mapper func(element T) R) Stream[R] {
	switch p := s.(type) {
	case *EmptyStream[T]:
		return &EmptyStream[R]{}
	case *SingletonStream[T]:
		return &SingletonStream[R]{mapper(p.Value)}
	case *pipeline[T]:
		return &pipeline[R]{
			opWrapSink: func(s sink[R]) {
				p.opWrapSink(mapWrapSink(s, mapper))
			},
		}
	}
	return nil
}

type iterator[T any] interface {
	advance(action func(T)) bool
}

type sliceIterator[T any] struct {
	x []T
}

func (it *sliceIterator[T]) advance(action func(T)) bool {
	if len(it.x) > 0 {
		action(it.x[0])
		it.x = it.x[1:]
	}
	return len(it.x) > 0
}

type sink[T any] interface {
	begin()
	end()
	accept(T)
}

type chainedSink[T, OUT any] struct {
	downstream sink[OUT]
	acceptFunc func(T)
}

func (cs *chainedSink[T, OUT]) begin() {
	cs.downstream.begin()
}

func (cs *chainedSink[T, OUT]) end() {
	cs.downstream.end()
}

func (cs *chainedSink[T, OUT]) accept(x T) {
	cs.acceptFunc(x)
}

func mapWrapSink[T, R any](s sink[R], mapper func(element T) R) sink[T] {
	return &chainedSink[T, R]{
		downstream: s,
		acceptFunc: func(x T) {
			s.accept(mapper(x))
		},
	}
}

func filterWrapSink[T any](s sink[T], predicate func(element T) bool) sink[T] {
	return &chainedSink[T, T]{
		downstream: s,
		acceptFunc: func(x T) {
			if predicate(x) {
				s.accept(x)
			}
		},
	}
}

type pipeline[OUT any] struct {
	opWrapSink func(sink[OUT])
}

func (p *pipeline[OUT]) Filter(predicate func(OUT) bool) Stream[OUT] {
	return &pipeline[OUT]{
		opWrapSink: func(s sink[OUT]) {
			p.opWrapSink(filterWrapSink(s, predicate))
		},
	}
}

func (p *pipeline[OUT]) Reduce(func(_, _ OUT) OUT) (OUT, bool) {
	var zero OUT
	return zero, false
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
