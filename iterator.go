package stream

type iterator[T any] interface {
	copyInto(sink[T])
}

type sliceIterator[T any] []T

func (it sliceIterator[T]) copyInto(s sink[T]) {
	s.begin()
	for _, v := range it {
		if !s.done() {
			s.accept(v)
		} else {
			break
		}
	}
	s.end()
}

type generatorIterator[T any] func() T

func (it generatorIterator[T]) copyInto(s sink[T]) {
	s.begin()
	for !s.done() {
		s.accept(it())
	}
	s.end()
}

type seedIterator[T any] struct {
	x        T
	operator func(T) T
}

func (it *seedIterator[T]) copyInto(s sink[T]) {
	s.begin()
	for !s.done() {
		s.accept(it.x)
		it.x = it.operator(it.x)
	}
	s.end()
}

type whileIterator[T any] struct {
	hasNext func() bool
	next    func() T
}

func (it *whileIterator[T]) copyInto(s sink[T]) {
	s.begin()
	for !s.done() && it.hasNext() {
		s.accept(it.next())
	}
	s.end()
}
