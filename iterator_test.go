package stream

import (
	"reflect"
	"testing"
)

func TestSliceIterator(t *testing.T) {
	in := []string{"a", "b", "c"}
	it := sliceIterator[string](in)

	ts := &testSink[string]{}

	it.copyInto(ts)

	if !reflect.DeepEqual(ts.result, in) {
		t.Errorf("wrong result, expected: %v, got: %v", in, ts.result)
	}
}
