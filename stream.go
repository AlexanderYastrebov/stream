package stream

type Stream[S, T any] interface {
	Filter(predicate func(element T) bool) Stream[S, T]
	Peek(consumer func(element T)) Stream[S, T]

	Limit(n int) Stream[S, T]
	Skip(n int) Stream[S, T]

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
	it := &sliceIterator[T]{x}
	return head[T](it)
}

func Map[S, T, R any](st Stream[S, T], mapper func(element T) R) Stream[S, R] {
	p, ok := st.(*pipeline[S, T])
	if !ok {
		panic("unexpected stream type")
	}
	return &pipeline[S, R]{
		wrapSink: func(s sink[R], done func(iterator[S], sink[S])) {
			p.wrapSink(mapWrapSink(s, mapper), done)
		},
	}
}

func Reduce[S, T, A any](st Stream[S, T], identity A, accumulator func(A, T) A) A {
	p, ok := st.(*pipeline[S, T])
	if !ok {
		panic("unexpected stream type")
	}

	a := &accumulatorSink[T, A]{value: identity, accumulator: accumulator}
	var aa sink[T] = a

	evaluate(p, aa)

	return a.value
}

func evaluate[S, T any](p *pipeline[S, T], out sink[T]) {
	var it iterator[S]
	var s sink[S]
	p.wrapSink(out, func(ii iterator[S], ss sink[S]) { it, s = ii, ss })
	s.begin()
	for !s.done() && it.advance(s.accept) {
	}
	s.end()
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

func peekWrapSink[T any](s sink[T], consumer func(element T)) sink[T] {
	return &chainedSink[T, T]{
		downstream: s,
		acceptFunc: func(x T) {
			consumer(x)
			s.accept(x)
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

type pipeline[S, OUT any] struct {
	wrapSink func(sink[OUT], func(iterator[S], sink[S]))
}

func head[S any](it iterator[S]) *pipeline[S, S] {
	return &pipeline[S, S]{
		wrapSink: func(s sink[S], done func(iterator[S], sink[S])) {
			done(it, s)
		},
	}
}

func (p *pipeline[S, OUT]) Filter(predicate func(OUT) bool) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		wrapSink: func(s sink[OUT], done func(iterator[S], sink[S])) {
			p.wrapSink(filterWrapSink(s, predicate), done)
		},
	}
}

func (p *pipeline[S, OUT]) Peek(consumer func(OUT)) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		wrapSink: func(s sink[OUT], done func(iterator[S], sink[S])) {
			p.wrapSink(peekWrapSink(s, consumer), done)
		},
	}
}

func (p *pipeline[S, OUT]) Limit(n int) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		wrapSink: func(s sink[OUT], done func(iterator[S], sink[S])) {
			p.wrapSink(limitWrapSink(s, n), done)
		},
	}
}

func (p *pipeline[S, OUT]) Skip(n int) Stream[S, OUT] {
	return &pipeline[S, OUT]{
		wrapSink: func(s sink[OUT], done func(iterator[S], sink[S])) {
			p.wrapSink(skipWrapSink(s, n), done)
		},
	}
}

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
	return func(x T) bool { return o.observe(x) }
}

func DistinctUsing[T any, C comparable](mapper func(T) C) func(T) bool {
	o := make(observer[C])
	return func(x T) bool { return o.observe(mapper(x)) }
}

type consumerSink[T any] struct {
	acceptFunc func(element T)
}

func (cs *consumerSink[T]) begin()     {}
func (cs *consumerSink[T]) end()       {}
func (cs *consumerSink[T]) done() bool { return false }
func (cs *consumerSink[T]) accept(x T) { cs.acceptFunc(x) }

func (p *pipeline[S, OUT]) ForEach(consumer func(OUT)) {
	cs := &consumerSink[OUT]{consumer}
	var s sink[OUT] = cs

	evaluate(p, s)
}

func (p *pipeline[S, OUT]) Reduce(accumulator func(a, b OUT) OUT) (OUT, bool) {
	var zero OUT
	foundAny := false
	return Reduce[S, OUT, OUT](p, zero, func(a OUT, e OUT) OUT {
		foundAny = true
		return accumulator(a, e)
	}), foundAny
}

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

func (p *pipeline[S, OUT]) AllMatch(predicate func(element OUT) bool) bool {
	ms := &matchSink[OUT]{predicate: predicate, stopWhen: false, stopValue: false}
	var s sink[OUT] = ms

	evaluate(p, s)

	return ms.value
}

func (p *pipeline[S, OUT]) AnyMatch(predicate func(element OUT) bool) bool {
	ms := &matchSink[OUT]{predicate: predicate, stopWhen: true, stopValue: true}
	var s sink[OUT] = ms

	evaluate(p, s)

	return ms.value
}

func (p *pipeline[S, OUT]) NoneMatch(predicate func(element OUT) bool) bool {
	ms := &matchSink[OUT]{predicate: predicate, stopWhen: true, stopValue: false}
	var s sink[OUT] = ms

	evaluate(p, s)

	return ms.value
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

func (p *pipeline[S, OUT]) FindFirst() (OUT, bool) {
	fs := &findSink[OUT]{}
	var s sink[OUT] = fs

	evaluate(p, s)

	return fs.value, fs.hasValue
}

func (p *pipeline[S, OUT]) Count() int {
	return Reduce[S, OUT, int](p, 0, func(a int, _ OUT) int { return a + 1 })
}

func (p *pipeline[S, OUT]) ToSlice() []OUT {
	return Reduce[S, OUT, []OUT](p, nil, func(a []OUT, e OUT) []OUT { return append(a, e) })
}
