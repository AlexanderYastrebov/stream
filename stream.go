package stream

type Stream[S, T any] interface {
	Filter(predicate func(element T) bool) Stream[S, T]
	Reduce(accumulator func(a, b T) T) (T, bool)
}

func Of[T any](x ...T) Stream[T, T] {
	return Slice(x)
}

func Slice[T any](x []T) Stream[T, T] {
	return nil
}

func Filter[S, T any](s Stream[S, T], predicate func(element T) bool) Stream[S, T] {
	return s.Filter(predicate)
}

func Map[S, T, R any](s Stream[S, T], mapper func(element T) R) Stream[S, R] {
	switch p := s.(type) {
	case *EmptyStream[S, T]:
		return &EmptyStream[S, R]{}
	case *SingletonStream[S, T]:
		return &SingletonStream[S, R]{mapper(p.Value)}
	case *pipeline[S, T]:
		return &pipeline[S, R]{
			opWrapSink: func(s sink[R], done func(iterator[S], sink[S])) {
				p.opWrapSink(mapWrapSink(s, mapper), done)
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

type pipeline[S, OUT any] struct {
	opWrapSink func(sink[OUT], func(iterator[S], sink[S]))
}

func (p *pipeline[S, OUT]) Filter(predicate func(OUT) bool) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		opWrapSink: func(s sink[OUT], done func(iterator[S], sink[S])) {
			p.opWrapSink(filterWrapSink(s, predicate), done)
		},
	}
}

func (p *pipeline[S, OUT]) Reduce(func(_, _ OUT) OUT) (OUT, bool) {
	var zero OUT
	return zero, false
}

func head[S any](it iterator[S]) *pipeline[S, S] {
	return &pipeline[S, S]{
		opWrapSink: func(s sink[S], done func(iterator[S], sink[S])) {
			done(it, s)
		},
	}
}

type EmptyStream[S, T any] struct{}

func (es *EmptyStream[S, T]) Filter(func(T) bool) Stream[S, T] {
	return es
}

func (es *EmptyStream[S, T]) Reduce(func(_, _ T) T) (T, bool) {
	var zero T
	return zero, false
}

type SingletonStream[S, T any] struct {
	Value T
}

func (s *SingletonStream[S, T]) Filter(predicate func(element T) bool) Stream[S, T] {
	if predicate(s.Value) {
		return s
	}
	return &EmptyStream[S, T]{}
}

func (s *SingletonStream[S, T]) Reduce(func(_, _ T) T) (T, bool) {
	return s.Value, true
}
