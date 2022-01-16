package stream

type Stream[S, T any] interface {
	Filter(predicate func(element T) bool) Stream[S, T]
	Reduce(accumulator func(a, b T) T) (T, bool)
}

func Of[T any](x ...T) Stream[T, T] {
	return Slice(x)
}

func Slice[T any](x []T) Stream[T, T] {
	it := &sliceIterator[T]{x}
	return head[T](it)
}

func Filter[S, T any](s Stream[S, T], predicate func(element T) bool) Stream[S, T] {
	return s.Filter(predicate)
}

func Map[S, T, R any](s Stream[S, T], mapper func(element T) R) Stream[S, R] {
	switch p := s.(type) {
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

type accumuatorSink[T any] struct {
	value       T
	foundAny    bool
	accumulator func(a, b T) T
}

func (as *accumuatorSink[T]) begin() {}

func (as *accumuatorSink[T]) end() {}

func (as *accumuatorSink[T]) accept(x T) {
	as.value = as.accumulator(as.value, x)
	as.foundAny = true
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

func (p *pipeline[S, OUT]) Reduce(accumulator func(a, b OUT) OUT) (OUT, bool) {
	var it iterator[S]
	var s sink[S]
	a := &accumuatorSink[OUT]{accumulator: accumulator}

	p.opWrapSink(a, func(ii iterator[S], ss sink[S]) { it, s = ii, ss })

	s.begin()
	for it.advance(s.accept) {
	}
	s.end()

	return a.value, a.foundAny
}

func head[S any](it iterator[S]) *pipeline[S, S] {
	return &pipeline[S, S]{
		opWrapSink: func(s sink[S], done func(iterator[S], sink[S])) {
			done(it, s)
		},
	}
}
