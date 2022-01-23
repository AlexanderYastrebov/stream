package stream

import "constraints"

type Stream[S, T any] interface {
	Filter(predicate func(element T) bool) Stream[S, T]
	Peek(consumer func(element T)) Stream[S, T]
	Limit(n int) Stream[S, T]
	Skip(n int) Stream[S, T]
	Sorted(less func(T, T) bool) Stream[S, T]
	ForEach(consumer func(element T))
	Reduce(accumulator func(a, b T) T) (T, bool)
	AllMatch(predicate func(element T) bool) bool
	AnyMatch(predicate func(element T) bool) bool
	NoneMatch(predicate func(element T) bool) bool
	FindFirst() (T, bool)
	Count() int
	ToSlice() []T
}

func Of[T any](x ...T) Stream[T, T] {
	return Slice(x)
}

func Slice[T any](x []T) Stream[T, T] {
	return head[T](&sliceIterator[T]{x})
}

func Generate[T any](generator func() T) Stream[T, T] {
	return head[T](generatorIterator[T](generator))
}

func Iterate[T any](seed T, operator func(T) T) Stream[T, T] {
	return head[T](&seedIterator[T]{seed, operator})
}

func While[T any](hasNext func() bool, supplier func() T) Stream[T, T] {
	return head[T](&whileIterator[T]{hasNext, supplier, hasNext()})
}

func Map[S, T, R any](st Stream[S, T], mapper func(element T) R) Stream[S, R] {
	p, ok := st.(*pipeline[S, T])
	if !ok {
		panic("unexpected stream type")
	}
	return &pipeline[S, R]{
		evaluate: func(s sink[R]) {
			p.evaluate(mapWrapSink(s, mapper))
		},
	}
}

func mapWrapSink[T, R any](s sink[R], mapper func(element T) R) sink[T] {
	return &chainedSink[T, R]{
		downstream: s,
		acceptFunc: func(x T) {
			s.accept(mapper(x))
		},
	}
}

func FlatMap[S, T, R any](st Stream[S, T], mapper func(element T) Stream[S, R]) Stream[S, R] {
	p, ok := st.(*pipeline[S, T])
	if !ok {
		panic("unexpected stream type")
	}
	return &pipeline[S, R]{
		evaluate: func(s sink[R]) {
			p.evaluate(flatMapWrapSink(s, mapper))
		},
	}
}

func flatMapWrapSink[S, T, R any](s sink[R], mapper func(element T) Stream[S, R]) sink[T] {
	return &chainedSink[T, R]{
		downstream: s,
		acceptFunc: func(x T) {
			mapper(x).ForEach(s.accept)
		},
	}
}

func Reduce[S, T, A any](st Stream[S, T], identity A, accumulator func(A, T) A) A {
	p, ok := st.(*pipeline[S, T])
	if !ok {
		panic("unexpected stream type")
	}

	a := &accumulatorSink[T, A]{value: identity, accumulator: accumulator}

	p.evaluate(a)

	return a.value
}

func NaturalOrder[T constraints.Ordered](a, b T) bool { return a < b }

func ReverseOrder[T constraints.Ordered](a, b T) bool { return a > b }

type observer[T comparable] map[T]struct{}

func (o observer[T]) observe(x T) bool {
	_, ok := o[x]
	if !ok {
		o[x] = struct{}{}
	}
	return !ok
}

func Distinct[T comparable]() func(T) bool {
	o := make(observer[T])
	return func(x T) bool {
		return o.observe(x)
	}
}

func DistinctUsing[T any, C comparable](mapper func(T) C) func(T) bool {
	o := make(observer[C])
	return func(x T) bool {
		return o.observe(mapper(x))
	}
}

type pipeline[S, OUT any] struct {
	evaluate func(sink[OUT])
}

func head[S any](it iterator[S]) *pipeline[S, S] {
	return &pipeline[S, S]{
		evaluate: func(s sink[S]) {
			copyInto(it, s)
		},
	}
}

func copyInto[T any](it iterator[T], s sink[T]) {
	s.begin()
	for !s.done() && it.advance(s.accept) {
	}
	s.end()
}

func (p *pipeline[S, OUT]) Filter(predicate func(OUT) bool) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		evaluate: func(s sink[OUT]) {
			p.evaluate(filterWrapSink(s, predicate))
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

func (p *pipeline[S, OUT]) Peek(consumer func(OUT)) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		evaluate: func(s sink[OUT]) {
			p.evaluate(peekWrapSink(s, consumer))
		},
	}
}

func peekWrapSink[T any](s sink[T], consumer func(element T)) sink[T] {
	return &chainedSink[T, T]{
		downstream: s,
		acceptFunc: func(x T) {
			consumer(x)
			s.accept(x)
		},
	}
}

func (p *pipeline[S, OUT]) Limit(n int) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		evaluate: func(s sink[OUT]) {
			p.evaluate(limitWrapSink(s, n))
		},
	}
}

func limitWrapSink[T any](s sink[T], n int) sink[T] {
	return &chainedSink[T, T]{
		downstream: s,
		doneFunc: func() bool {
			return n <= 0 || s.done()
		},
		acceptFunc: func(x T) {
			if n > 0 {
				s.accept(x)
				n--
			}
		},
	}
}

func (p *pipeline[S, OUT]) Skip(n int) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		evaluate: func(s sink[OUT]) {
			p.evaluate(skipWrapSink(s, n))
		},
	}
}

func skipWrapSink[T any](s sink[T], n int) sink[T] {
	return &chainedSink[T, T]{
		downstream: s,
		acceptFunc: func(x T) {
			if n > 0 {
				n--
				return
			}
			s.accept(x)
		},
	}
}

func (p *pipeline[S, OUT]) Sorted(less func(OUT, OUT) bool) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		evaluate: func(s sink[OUT]) {
			p.evaluate(&sortedSink[OUT]{downstream: s, less: less})
		},
	}
}

func (p *pipeline[S, OUT]) ForEach(consumer func(OUT)) {
	p.evaluate(consumerSink[OUT](consumer))
}

func (p *pipeline[S, OUT]) Reduce(accumulator func(a, b OUT) OUT) (OUT, bool) {
	var zero OUT
	foundAny := false
	return Reduce[S, OUT, OUT](p, zero, func(a OUT, e OUT) OUT {
		foundAny = true
		return accumulator(a, e)
	}), foundAny
}

func (p *pipeline[S, OUT]) AllMatch(predicate func(element OUT) bool) bool {
	s := &matchSink[OUT]{predicate: predicate, stopWhen: false, stopValue: false}

	p.evaluate(s)

	return s.value
}

func (p *pipeline[S, OUT]) AnyMatch(predicate func(element OUT) bool) bool {
	s := &matchSink[OUT]{predicate: predicate, stopWhen: true, stopValue: true}

	p.evaluate(s)

	return s.value
}

func (p *pipeline[S, OUT]) NoneMatch(predicate func(element OUT) bool) bool {
	s := &matchSink[OUT]{predicate: predicate, stopWhen: true, stopValue: false}

	p.evaluate(s)

	return s.value
}

func (p *pipeline[S, OUT]) FindFirst() (OUT, bool) {
	s := &findSink[OUT]{}

	p.evaluate(s)

	return s.value, s.hasValue
}

func (p *pipeline[S, OUT]) Count() int {
	return Reduce[S, OUT, int](p, 0, func(a int, _ OUT) int { return a + 1 })
}

func (p *pipeline[S, OUT]) ToSlice() []OUT {
	return Reduce[S, OUT, []OUT](p, nil, func(a []OUT, e OUT) []OUT { return append(a, e) })
}
