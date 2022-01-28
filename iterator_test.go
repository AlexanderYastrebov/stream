package stream

import (
	"reflect"
	"testing"
)

func TestSliceIterator(t *testing.T) {
	in := []string{"a", "b", "c"}
	ts := &testSink[string]{}

	sliceIterator[string](in).copyInto(ts)

	if !reflect.DeepEqual(ts.result, in) {
		t.Errorf("wrong result, expected: %v, got: %v", in, ts.result)
	}

	in = []string{"a", "b"}
	ts = &testSink[string]{}
	ts.setLimit(2)

	sliceIterator[string](in).copyInto(ts)

	if !reflect.DeepEqual(ts.result, in) {
		t.Errorf("wrong result, expected: %v, got: %v", in, ts.result)
	}
}
