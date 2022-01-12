package stream_test

import (
	"fmt"
	"github.com/AlexanderYastrebov/stream"
	"strings"
	"testing"
)

type prefixMatcher struct {
	prefix string
}

func (m *prefixMatcher) matches(s string) bool {
	return strings.HasPrefix(s, m.prefix)
}

func isHe(s string) bool {
	return strings.HasPrefix(s, "he")
}

func length(s string) int {
	return len(s)
}

func TestStream(t *testing.T) {
	s := stream.Of("hello", "world")
	t.Log(s)

	s = &stream.EmptyStream[string]{}

	t.Logf("%p", s)
	t.Logf("%p", s.Filter(func(_ string) bool { return true }))

	s = &stream.SingletonStream[string]{"hello1"}
	m := &prefixMatcher{"hello"}
	var z int
	var ok bool
	{
		x := s.
			Filter(isHe).
			Filter(m.matches).
			Filter(func(e string) bool { return true })
		y := stream.Map(x, length).
			Filter(func(i int) bool { return true })
		z, ok = y.Reduce(func(a, b int) int { return a + b })
	}
	t.Logf("%#v %t", z, ok)

	var a stream.Stream[string] = &stream.SingletonStream[string]{"hello world"}
	a = stream.Filter(a, isHe)
	a = stream.Filter(a, m.matches)
	a = stream.Filter(a, func(e string) bool { return true })
	b := stream.Map(a, length)
	b = stream.Filter(b, func(i int) bool { return i > 0 })
	b = stream.Filter(b, func(i int) bool { return true })
	c, ok := b.Reduce(func(i, j int) int { return i + j })

	t.Logf("%#v %t", c, ok)

}

func ExampleStream() {
	fmt.Printf("hello")
	// Output: hello
}
