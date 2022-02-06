package stream

import "sort"

type sink[T any] interface {
	begin()
	done() bool
	accept(T)
	end()
}

type chainedSink[IN, OUT any] struct {
	sink[OUT]
	doneFunc   func() bool
	acceptFunc func(IN)
}

func (s *chainedSink[IN, OUT]) done() bool {
	if s.doneFunc != nil {
		return s.doneFunc()
	}
	return s.sink.done()
}

func (s *chainedSink[IN, OUT]) accept(x IN) {
	s.acceptFunc(x)
}

func filterSink[T any](s sink[T], predicate func(element T) bool) sink[T] {
	return &chainedSink[T, T]{
		sink: s,
		acceptFunc: func(x T) {
			if predicate(x) {
				s.accept(x)
			}
		},
	}
}

func peekSink[T any](s sink[T], consumer func(element T)) sink[T] {
	return &chainedSink[T, T]{
		sink: s,
		acceptFunc: func(x T) {
			consumer(x)
			s.accept(x)
		},
	}
}

func limitSink[T any](s sink[T], n int) sink[T] {
	return &chainedSink[T, T]{
		sink: s,
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

func skipSink[T any](s sink[T], n int) sink[T] {
	return &chainedSink[T, T]{
		sink: s,
		acceptFunc: func(x T) {
			if n > 0 {
				n--
				return
			}
			s.accept(x)
		},
	}
}

func mapSink[T, R any](s sink[R], mapper func(element T) R) sink[T] {
	return &chainedSink[T, R]{
		sink: s,
		acceptFunc: func(x T) {
			s.accept(mapper(x))
		},
	}
}

func flatMapSink[T, R any](s sink[R], mapper func(element T) Stream[R]) sink[T] {
	return &chainedSink[T, R]{
		sink: s,
		acceptFunc: func(x T) {
			mapper(x).ForEach(s.accept)
		},
	}
}

type base struct{}

func (base) begin()     {}
func (base) done() bool { return false }
func (base) end()       {}

type accumulatorSink[T, A any] struct {
	base
	value       A
	accumulator func(a A, b T)
}

func (s *accumulatorSink[T, A]) accept(x T) {
	s.accumulator(s.value, x)
}

type sortedSink[T any] struct {
	base
	downstream sink[T]
	less       func(T, T) bool
	slice      []T
}

func (s *sortedSink[T]) accept(x T) {
	s.slice = append(s.slice, x)
}

func (s *sortedSink[T]) end() {
	sort.SliceStable(s.slice, func(i, j int) bool {
		return s.less(s.slice[i], s.slice[j])
	})
	sliceIterator[T](s.slice).copyInto(s.downstream)
	s.slice = nil
}

type consumerSink[T any] func(T)

func (s consumerSink[T]) begin()     {}
func (s consumerSink[T]) done() bool { return false }
func (s consumerSink[T]) accept(x T) { s(x) }
func (s consumerSink[T]) end()       {}

type matchSink[T any] struct {
	base
	predicate func(element T) bool
	stopWhen  bool
	stopValue bool

	value    bool
	hasValue bool
}

func (s *matchSink[T]) begin() {
	s.value = !s.stopValue
}

func (s *matchSink[T]) done() bool {
	return s.hasValue
}

func (s *matchSink[T]) accept(x T) {
	if !s.hasValue && s.predicate(x) == s.stopWhen {
		s.value = s.stopValue
		s.hasValue = true
	}
}

type findSink[T any] struct {
	base
	value    T
	hasValue bool
}

func (s *findSink[T]) done() bool {
	return s.hasValue
}

func (s *findSink[T]) accept(x T) {
	if !s.hasValue {
		s.value = x
		s.hasValue = true
	}
}

type minSink[T any] struct {
	base
	less     func(T, T) bool
	value    T
	hasValue bool
}

func (s *minSink[T]) accept(x T) {
	if !s.hasValue {
		s.value = x
		s.hasValue = true
	} else if s.less(x, s.value) {
		s.value = x
	}
}

type maxSink[T any] struct {
	base
	less     func(T, T) bool
	value    T
	hasValue bool
}

func (s *maxSink[T]) accept(x T) {
	if !s.hasValue {
		s.value = x
		s.hasValue = true
	} else if s.less(s.value, x) {
		s.value = x
	}
}

type forwardingSink[T any] struct {
	sink[T]
}

func (forwardingSink[T]) begin() {}
func (forwardingSink[T]) end()   {}
