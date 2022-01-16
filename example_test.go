package stream_test

import (
	"fmt"
	_ "github.com/AlexanderYastrebov/stream"
	"strings"
	_ "testing"
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

func ExampleStream() {
	fmt.Printf("hello")
	// Output: hello
}
