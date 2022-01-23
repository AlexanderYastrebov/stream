package stream

import "constraints"

type Stream[T any] interface {
	Filter(predicate func(element T) bool) Stream[T]
	Peek(consumer func(element T)) Stream[T]
	Limit(n int) Stream[T]
	Skip(n int) Stream[T]
	Sorted(less func(T, T) bool) Stream[T]
	ForEach(consumer func(element T))
	Reduce(accumulator func(a, b T) T) (T, bool)
	AllMatch(predicate func(element T) bool) bool
	AnyMatch(predicate func(element T) bool) bool
	NoneMatch(predicate func(element T) bool) bool
	FindFirst() (T, bool)
	Min(less func(T, T) bool) (T, bool)
	Max(less func(T, T) bool) (T, bool)
	Count() int
	ToSlice() []T
}

func Of[T any](x ...T) Stream[T] {
	return Slice(x)
}

func Slice[T any](x []T) Stream[T] {
	return head[T](&sliceIterator[T]{x})
}

func Generate[T any](generator func() T) Stream[T] {
	return head[T](generatorIterator[T](generator))
}

func Iterate[T any](seed T, operator func(T) T) Stream[T] {
	return head[T](&seedIterator[T]{seed, operator})
}

func While[T any](hasNext func() bool, supplier func() T) Stream[T] {
	return head[T](&whileIterator[T]{hasNext, supplier, hasNext()})
}

func Map[T, R any](st Stream[T], mapper func(element T) R) Stream[R] {
	p, ok := st.(*pipeline[T])
	if !ok {
		panic("unexpected stream type")
	}
	return &pipeline[R]{
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

func FlatMap[T, R any](st Stream[T], mapper func(element T) Stream[R]) Stream[R] {
	p, ok := st.(*pipeline[T])
	if !ok {
		panic("unexpected stream type")
	}
	return &pipeline[R]{
		evaluate: func(s sink[R]) {
			p.evaluate(flatMapWrapSink(s, mapper))
		},
	}
}

func flatMapWrapSink[T, R any](s sink[R], mapper func(element T) Stream[R]) sink[T] {
	return &chainedSink[T, R]{
		downstream: s,
		acceptFunc: func(x T) {
			mapper(x).ForEach(s.accept)
		},
	}
}

func Reduce[T, A any](st Stream[T], identity A, accumulator func(A, T) A) A {
	p, ok := st.(*pipeline[T])
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

type pipeline[T any] struct {
	evaluate func(sink[T])
}

func head[T any](it iterator[T]) *pipeline[T] {
	return &pipeline[T]{
		evaluate: func(s sink[T]) {
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

func (p *pipeline[T]) Filter(predicate func(T) bool) Stream[T] {
	return &pipeline[T]{
		evaluate: func(s sink[T]) {
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

func (p *pipeline[T]) Peek(consumer func(T)) Stream[T] {
	return &pipeline[T]{
		evaluate: func(s sink[T]) {
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

func (p *pipeline[T]) Limit(n int) Stream[T] {
	return &pipeline[T]{
		evaluate: func(s sink[T]) {
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

func (p *pipeline[T]) Skip(n int) Stream[T] {
	return &pipeline[T]{
		evaluate: func(s sink[T]) {
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

func (p *pipeline[T]) Sorted(less func(T, T) bool) Stream[T] {
	return &pipeline[T]{
		evaluate: func(s sink[T]) {
			p.evaluate(&sortedSink[T]{downstream: s, less: less})
		},
	}
}

func (p *pipeline[T]) ForEach(consumer func(T)) {
	p.evaluate(consumerSink[T](consumer))
}

func (p *pipeline[T]) Reduce(accumulator func(a, b T) T) (T, bool) {
	var zero T
	foundAny := false
	return Reduce[T, T](p, zero, func(a T, e T) T {
		foundAny = true
		return accumulator(a, e)
	}), foundAny
}

func (p *pipeline[T]) AllMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: false, stopValue: false}

	p.evaluate(s)

	return s.value
}

func (p *pipeline[T]) AnyMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: true, stopValue: true}

	p.evaluate(s)

	return s.value
}

func (p *pipeline[T]) NoneMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: true, stopValue: false}

	p.evaluate(s)

	return s.value
}

func (p *pipeline[T]) FindFirst() (T, bool) {
	s := &findSink[T]{}

	p.evaluate(s)

	return s.value, s.hasValue
}

func (p *pipeline[T]) Min(less func(T, T) bool) (T, bool) {
	s := &minSink[T]{less: less}

	p.evaluate(s)

	return s.value, s.hasValue
}

func (p *pipeline[T]) Max(less func(T, T) bool) (T, bool) {
	s := &maxSink[T]{less: less}

	p.evaluate(s)

	return s.value, s.hasValue
}

func (p *pipeline[T]) Count() int {
	return Reduce[T, int](p, 0, func(a int, _ T) int { return a + 1 })
}

func (p *pipeline[T]) ToSlice() []T {
	return Reduce[T, []T](p, nil, func(a []T, e T) []T { return append(a, e) })
}
