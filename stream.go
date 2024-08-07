package stream

type Stream[T any] interface {
	Filter(predicate func(element T) bool) Stream[T]
	Map(mapper func(element T) T) Stream[T]
	FlatMap(mapper func(element T) Stream[T]) Stream[T]
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
	return head(sliceIterator[T](x))
}

func Generate[T any](generator func() T) Stream[T] {
	return head(generatorIterator[T](generator))
}

func Iterate[T any](seed T, operator func(T) T) Stream[T] {
	return head(&seedIterator[T]{seed, operator})
}

func While[T any](hasNext func() bool, supplier func() T) Stream[T] {
	return head(&whileIterator[T]{hasNext, supplier})
}

type stage[T any] func(sink[T])

func head[T any](it iterator[T]) Stream[T] {
	return stage[T](it.copyInto)
}

func (p stage[T]) copyInto(s sink[T]) {
	p(s)
}

func (p stage[T]) Filter(predicate func(T) bool) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(filterSink(s, predicate))
	})
}

func (p stage[T]) Map(mapper func(T) T) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(mapSink(s, mapper))
	})
}

func (p stage[T]) FlatMap(mapper func(T) Stream[T]) Stream[T] {
	return stage[T](func(s sink[T]) {
		p.copyInto(flatMapSink(s, mapper))
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
		as := appendSink[T]{s}
		s.begin()
		p.copyInto(as)
		if !as.done() {
			st.copyInto(as)
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
	return p.Reduce(func(r, x T) T {
		if less(x, r) {
			return x
		}
		return r
	})
}

func (p stage[T]) Max(less func(T, T) bool) (T, bool) {
	return p.Reduce(func(r, x T) T {
		if less(r, x) {
			return x
		}
		return r
	})
}

func (p stage[T]) Count() (result int) {
	p.ForEach(func(x T) {
		result++
	})
	return
}

func (p stage[T]) ToSlice() (result []T) {
	p.ForEach(func(x T) {
		result = append(result, x)
	})
	return
}
