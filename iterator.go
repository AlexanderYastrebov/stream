package stream

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

type generatorIterator[T any] func() T

func (it generatorIterator[T]) advance(action func(T)) bool {
	action(it())
	return true
}

type seedIterator[T any] struct {
	x        T
	operator func(T) T
}

func (it *seedIterator[T]) advance(action func(T)) bool {
	action(it.x)
	it.x = it.operator(it.x)
	return true
}

type whileIterator[T any] struct {
	check   func() bool
	next    func() T
	hasNext bool
}

func (it *whileIterator[T]) advance(action func(T)) bool {
	if it.hasNext {
		action(it.next())
		it.hasNext = it.check()
	}
	return it.hasNext
}
