package stream

import (
	"fmt"
	"testing"
)

type testSink[T any] struct {
	limit        int
	began, ended bool
	result       []T
}

func (s *testSink[T]) begin() {
	s.began = true
}

func (s *testSink[T]) done() bool {
	return s.limit <= 0
}

func (s *testSink[T]) accept(x T) {
	s.result = append(s.result, x)
	s.limit--
}

func (s *testSink[T]) end() {
	s.ended = true
}

func TestChainedSink(t *testing.T) {
	var s sink[string]

	s = &chainedSink[string, string]{
		acceptFunc: func(x string) {
			t.Logf("sink: %s", x)
		},
	}
	q := mapSink(s, func(i int) string { return fmt.Sprintf(">>%d<<", i) })
	r := filterSink(q, func(i int) bool { return i > 1 })

	r.accept(10)
}
