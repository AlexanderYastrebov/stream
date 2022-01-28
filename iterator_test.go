package stream

import (
	"reflect"
	"testing"
)

func TestSliceIterator(t *testing.T) {
	in := []string{"a", "b", "c"}
	it := sliceIterator[string](in)

	var res []string
	it.copyInto(consumerSink[string](func(s string) { res = append(res, s) }))

	if !reflect.DeepEqual(res, in) {
		t.Errorf("wrong result, expected: %v, got: %v", in, res)
	}
}
