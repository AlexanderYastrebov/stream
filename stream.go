package stream

type Stream[T any] func(sink[T])

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

func head[T any](it iterator[T]) Stream[T] {
	return Stream[T](it.copyInto)
}

func (p Stream[T]) copyInto(s sink[T]) {
	p(s)
}

func (p Stream[T]) Filter(predicate func(T) bool) Stream[T] {
	return Stream[T](func(s sink[T]) {
		p.copyInto(filterSink(s, predicate))
	})
}

func (p Stream[T]) Map[R any](mapper func(element T) R) Stream[R] {
	return Stream[R](func(s sink[R]) {
		p.copyInto(mapSink(s, mapper))
	})
}

func (p Stream[T]) FlatMap[R any](mapper func(T) Stream[R]) Stream[R] {
	return Stream[R](func(s sink[R]) {
		p.copyInto(flatMapSink(s, mapper))
	})
}

func (p Stream[T]) Collect[A any](identity A, accumulator func(A, T)) A {
	a := &accumulatorSink[T, A]{value: identity, accumulator: accumulator}

	p.copyInto(a)

	return a.value
}

func (p Stream[T]) Peek(consumer func(T)) Stream[T] {
	return Stream[T](func(s sink[T]) {
		p.copyInto(peekSink(s, consumer))
	})
}

func (p Stream[T]) Limit(n int) Stream[T] {
	return Stream[T](func(s sink[T]) {
		p.copyInto(limitSink(s, n))
	})
}

func (p Stream[T]) Skip(n int) Stream[T] {
	return Stream[T](func(s sink[T]) {
		p.copyInto(skipSink(s, n))
	})
}

func (p Stream[T]) Sort(cmp func(T, T) int) Stream[T] {
	return Stream[T](func(s sink[T]) {
		p.copyInto(&sortSink[T]{downstream: s, cmp: cmp})
	})
}

func (p Stream[T]) Append(st Stream[T]) Stream[T] {
	return Stream[T](func(s sink[T]) {
		as := appendSink[T]{s}
		s.begin()
		p.copyInto(as)
		if !as.done() {
			st.copyInto(as)
		}
		s.end()
	})
}

func (p Stream[T]) ForEach(consumer func(T)) {
	p.copyInto(consumerSink[T](consumer))
}

func (p Stream[T]) Reduce(combiner func(T, T) T) (T, bool) {
	var result T
	foundAny := false
	p.ForEach(func(x T) {
		if !foundAny {
			foundAny = true
			result = x
		} else {
			result = combiner(result, x)
		}
	})
	return result, foundAny
}

func (p Stream[T]) AllMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: false, stopValue: false}

	p.copyInto(s)

	return s.value
}

func (p Stream[T]) AnyMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: true, stopValue: true}

	p.copyInto(s)

	return s.value
}

func (p Stream[T]) NoneMatch(predicate func(element T) bool) bool {
	s := &matchSink[T]{predicate: predicate, stopWhen: true, stopValue: false}

	p.copyInto(s)

	return s.value
}

func (p Stream[T]) FindFirst() (T, bool) {
	s := &findSink[T]{}

	p.copyInto(s)

	return s.value, s.hasValue
}

func (p Stream[T]) Min(cmp func(T, T) int) (T, bool) {
	return p.Reduce(func(r, x T) T {
		c := cmp(x, r)
		if c < 0 {
			return x
		}
		return r
	})
}

func (p Stream[T]) Max(cmp func(T, T) int) (T, bool) {
	return p.Reduce(func(r, x T) T {
		c := cmp(x, r)
		if c > 0 {
			return x
		}
		return r
	})
}

func (p Stream[T]) Count() (result int) {
	p.ForEach(func(x T) {
		result++
	})
	return
}

func (p Stream[T]) ToSlice() (result []T) {
	p.ForEach(func(x T) {
		result = append(result, x)
	})
	return
}
