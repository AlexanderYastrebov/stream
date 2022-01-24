package stream

import (
	"fmt"
	"testing"
)

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
