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
