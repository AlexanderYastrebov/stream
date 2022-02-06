package stream

import "constraints"

type Stream[T any] interface {
	Filter(predicate func(element T) bool) Stream[T]
	Peek(consumer func(element T)) Stream[T]
	Limit(n int) Stream[T]
	Skip(n int) Stream[T]
	Sorted(less func(T, T) bool) Stream[T]
	Append(Stream[T]) Stream[T]

	ForEach(consumer func(element T))
	Reduce(accumulator func(T, T) T) (T, bool)
	AllMatch(predicate func(element T) bool) bool
	AnyMatch(predicate func(element T) bool) bool
	NoneMatch(predicate func(element T) bool) bool
	FindFirst() (T, bool)
	Min(less func(T, T) bool) (T, bool)
	Max(less func(T, T) bool) (T, bool)
	Count() int
	ToSlice() []T

	copyInto(sink[T])
}

func Map[T, R any](st Stream[T], mapper func(element T) R) Stream[R] {
	return stage[R](func(s sink[R]) {
		st.copyInto(mapSink(s, mapper))
	})
}

func FlatMap[T, R any](st Stream[T], mapper func(element T) Stream[R]) Stream[R] {
	return stage[R](func(s sink[R]) {
		st.copyInto(flatMapSink(s, mapper))
	})
}

func Collect[T, A any](st Stream[T], identity A, accumulator func(A, T)) A {
	a := &accumulatorSink[T, A]{value: identity, accumulator: accumulator}

	st.copyInto(a)

	return a.value
}

func Of[T any](x ...T) Stream[T] {
	return Slice(x)
}

func Slice[T any](x []T) Stream[T] {
	return head[T](sliceIterator[T](x))
}

func Generate[T any](generator func() T) Stream[T] {
	return head[T](generatorIterator[T](generator))
}

func Iterate[T any](seed T, operator func(T) T) Stream[T] {
	return head[T](&seedIterator[T]{seed, operator})
}

func While[T any](hasNext func() bool, supplier func() T) Stream[T] {
	return head[T](&whileIterator[T]{hasNext, supplier})
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

type stage[T any] func(sink[T])

func head[T any](it iterator[T]) Stream[T] {
	return stage[T](func(s sink[T]) {
		it.copyInto(s)
	})
}

func (p stage[T]) copyInto(s sink[T]) {
	p(s)
}

func (p stage[T]) Filter(predicate func(T) bool) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(filterSink(s, predicate))
	})
}

func (p stage[T]) Peek(consumer func(T)) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(peekSink(s, consumer))
	})
}

func (p stage[T]) Limit(n int) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(limitSink(s, n))
	})
}

func (p stage[T]) Skip(n int) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(skipSink(s, n))
	})
}

func (p stage[T]) Sorted(less func(T, T) bool) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(&sortedSink[T]{downstream: s, less: less})
	})
}

func (p stage[T]) Append(st Stream[T]) Stream[T] {
	return stage[T](func(s sink[T]) {
		fs := forwardingSink[T]{s}
		s.begin()
		p.copyInto(fs)
		if !s.done() {
			st.copyInto(fs)
		}
		s.end()
	})
}

func (p stage[T]) ForEach(consumer func(T)) {
	p.copyInto(consumerSink[T](consumer))
}

func (p stage[T]) Reduce(accumulator func(T, T) T) (T, bool) {
	var result T
	foundAny := false
	p.ForEach(func(x T) {
		if !foundAny {
			foundAny = true
			result = x
		} else {
			result = accumulator(result, x)
		}
	})
	return result, foundAny
}

func (p stage[T]) AllMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: false, stopValue: false}

	p.copyInto(s)

	return s.value
}

func (p stage[T]) AnyMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: true, stopValue: true}

	p.copyInto(s)

	return s.value
}

func (p stage[T]) NoneMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: true, stopValue: false}

	p.copyInto(s)

	return s.value
}

func (p stage[T]) FindFirst() (T, bool) {
	s := &findSink[T]{}

	p.copyInto(s)

	return s.value, s.hasValue
}

func (p stage[T]) Min(less func(T, T) bool) (T, bool) {
	s := &minSink[T]{less: less}

	p.copyInto(s)

	return s.value, s.hasValue
}

func (p stage[T]) Max(less func(T, T) bool) (T, bool) {
	s := &maxSink[T]{less: less}

	p.copyInto(s)

	return s.value, s.hasValue
}

func (p stage[T]) Count() (result int) {
	p.ForEach(func(x T) { result++ })
	return
}

func (p stage[T]) ToSlice() (result []T) {
	p.ForEach(func(x T) { result = append(result, x) })
	return
}
