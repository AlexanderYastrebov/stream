package stream

import "sort"

type sink[T any] interface {
	begin()
	end()
	done() bool
	accept(T)
}

type chainedSink[T, OUT any] struct {
	downstream sink[OUT]
	doneFunc   func() bool
	acceptFunc func(T)
}

func (cs *chainedSink[T, OUT]) begin() {
	cs.downstream.begin()
}

func (cs *chainedSink[T, OUT]) end() {
	cs.downstream.end()
}

func (cs *chainedSink[T, OUT]) done() bool {
	if cs.doneFunc != nil {
		return cs.doneFunc()
	}
	return cs.downstream.done()
}

func (cs *chainedSink[T, OUT]) accept(x T) {
	cs.acceptFunc(x)
}

type accumulatorSink[T, A any] struct {
	value       A
	accumulator func(a A, b T) A
}

func (as *accumulatorSink[T, A]) begin()     {}
func (as *accumulatorSink[T, A]) end()       {}
func (as *accumulatorSink[T, A]) done() bool { return false }

func (as *accumulatorSink[T, A]) accept(x T) {
	as.value = as.accumulator(as.value, x)
}

type sortedSink[T any] struct {
	downstream sink[T]
	less       func(T, T) bool
	slice      []T
}

func (s *sortedSink[T]) begin()     {}
func (s *sortedSink[T]) done() bool { return false }

func (s *sortedSink[T]) end() {
	sort.SliceStable(s.slice, func(i, j int) bool {
		return s.less(s.slice[i], s.slice[j])
	})
	var it iterator[T] = &sliceIterator[T]{s.slice}
	copyInto(it, s.downstream)
	s.slice = nil
}

func (s *sortedSink[T]) accept(x T) {
	s.slice = append(s.slice, x)
}

type consumerSink[T any] func(T)

func (cs consumerSink[T]) begin()     {}
func (cs consumerSink[T]) end()       {}
func (cs consumerSink[T]) done() bool { return false }
func (cs consumerSink[T]) accept(x T) { cs(x) }

type matchSink[T any] struct {
	predicate func(element T) bool
	stopWhen  bool
	stopValue bool

	value    bool
	hasValue bool
}

func (ms *matchSink[T]) begin()     { ms.value = !ms.stopValue }
func (ms *matchSink[T]) end()       {}
func (ms *matchSink[T]) done() bool { return ms.hasValue }

func (ms *matchSink[T]) accept(x T) {
	if !ms.hasValue && ms.predicate(x) == ms.stopWhen {
		ms.value = ms.stopValue
		ms.hasValue = true
	}
}

type findSink[T any] struct {
	value    T
	hasValue bool
}

func (fs *findSink[T]) begin()     {}
func (fs *findSink[T]) end()       {}
func (fs *findSink[T]) done() bool { return fs.hasValue }

func (fs *findSink[T]) accept(x T) {
	if !fs.hasValue {
		fs.value = x
		fs.hasValue = true
	}
}
