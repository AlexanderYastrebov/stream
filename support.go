package stream

import "golang.org/x/exp/constraints"

func NaturalOrder[T constraints.Ordered](a, b T) bool {
	return a < b
}

func ReverseOrder[T constraints.Ordered](a, b T) bool {
	return a > b
}

func Distinct[T comparable]() func(T) bool {
	o := make(observer[T])
	return o.observe
}

func DistinctUsing[T any, C comparable](mapper func(T) C) func(T) bool {
	o := make(observer[C])
	return func(x T) bool {
		return o.observe(mapper(x))
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
