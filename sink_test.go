package stream

import (
	"fmt"
	"testing"
)

type testSink[T any] struct {
	count        int
	limit        int
	began, ended bool
	log          []string
}

func (s *testSink[T]) begin() {
	s.log = append(s.log, "begin")
}

func (s *testSink[T]) done() bool {
	if s.limit < 0 {
		return false
	}
	return s.count >= s.limit
}

func (s *testSink[T]) accept(x T) {
	s.log = append(s.log, fmt.Sprintf("accept(%v)", x))
	s.count++
}

func (s *testSink[T]) end() {
	s.log = append(s.log, "end")
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
