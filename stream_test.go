package stream

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSliceIterator(t *testing.T) {
	in := []string{"a", "b", "c"}
	it := &sliceIterator[string]{in}

	var res []string
	for it.advance(func(s string) { res = append(res, s) }) {
	}

	if !reflect.DeepEqual(res, in) {
		t.Errorf("wrong result, expected: %v, got: %v", in, res)
	}
}

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
	p := head[string](nil)

	q := p.
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

	rr.opWrapSink(s, func(_ iterator[string], s sink[string]) {
		x = s
	})

	x.accept("test1")
	x.accept("test22")
}

func TestReduce(t *testing.T) {
	in := []string{"a", "bb", "ccc", "dddd"}
	it := &sliceIterator[string]{in}
	p := head[string](it)

	q := p.
		Filter(func(s string) bool { return len(s) > 1 })

	r := Map(q, func(s string) int { return len(s) })

	x, ok := r.Reduce(func(a, b int) int { return a + b })

	t.Log(x, ok)
	if x != 9 {
		t.Errorf("expected 9, got: %d", x)
	}
	if !ok {
		t.Errorf("expected ok")
	}
}
