package stream

import (
	"fmt"
	"strings"
	"testing"
)

func TestStream(t *testing.T) {
	e := &emptyStream[string]{}

	t.Logf("%p", e)
	t.Logf("%p", e.Filter(func(_ string) bool { return true }))

	s := &singletonStream[string]{"hello"}

	t.Logf("%p", s)
	t.Logf("%p", s.Filter(func(_ string) bool { return true }))
	t.Logf("%p", s.Filter(func(e string) bool { return strings.HasPrefix(e, "test") }))
}

func ExampleStream() {
	fmt.Printf("hello")
	// Output: hello
}
