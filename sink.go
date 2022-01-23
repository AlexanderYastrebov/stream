package stream

import "sort"

type sink[T any] interface {
	begin()
	done() bool
	accept(T)
	end()
}

type chainedSink[T, OUT any] struct {
	downstream sink[OUT]
	doneFunc   func() bool
	acceptFunc func(T)
}

func (s *chainedSink[T, OUT]) begin() {
	s.downstream.begin()
}

func (s *chainedSink[T, OUT]) done() bool {
	if s.doneFunc != nil {
		return s.doneFunc()
	}
	return s.downstream.done()
}

func (s *chainedSink[T, OUT]) accept(x T) {
	s.acceptFunc(x)
}

func (s *chainedSink[T, OUT]) end() {
	s.downstream.end()
}

type base struct{}

func (base) begin()     {}
func (base) done() bool { return false }
func (base) end()       {}

type accumulatorSink[T, A any] struct {
	base
	value       A
	accumulator func(a A, b T) A
}

func (s *accumulatorSink[T, A]) accept(x T) {
	s.value = s.accumulator(s.value, x)
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
	var it iterator[T] = &sliceIterator[T]{s.slice}
	copyInto(it, s.downstream)
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
