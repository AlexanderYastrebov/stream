package stream

import (
	"fmt"
	"testing"
)

func TestChainedSink(t *testing.T) {
	var s sink[string]

	s = &chainedSink[string, string]{
		downstream: nil,
		acceptFunc: func(x string) {
			t.Logf("sink: %s", x)
		},
	}
	q := mapWrapSink(s, func(i int) string { return fmt.Sprintf(">>%d<<", i) })
	r := filterWrapSink(q, func(i int) bool { return i > 1 })

	r.accept(10)
}

func TestPipeline(t *testing.T) {
	q := Of("_").
		Filter(func(s string) bool { return true })

	r := Map(q, func(s string) int { return len(s) }).
		Filter(func(i int) bool { return true })

	u := Map(r, func(i int) float64 { return float64(i) })

	s := &chainedSink[float64, float64]{
		downstream: nil,
		acceptFunc: func(x float64) {
			t.Logf("accept: %f", x)
		},
	}

	rr := u.(*pipeline[string, float64])

	var x sink[string]

	rr.wrapSink(s, func(_ iterator[string], s sink[string]) {
		x = s
	})

	x.accept("test1")
	x.accept("test22")
}
