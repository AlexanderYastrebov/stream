package stream

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIterators(t *testing.T) {
	for _, tc := range []struct {
		iterator iterator[string]
		limit    int
		expect   []string
	}{
		{
			iterator: sliceIterator[string]([]string{"a", "b", "c"}),
			limit:    -1,
			expect:   []string{"begin", "accept(a)", "accept(b)", "accept(c)", "end"},
		},
		{
			iterator: sliceIterator[string]([]string{"a", "b", "c"}),
			limit:    0,
			expect:   []string{"begin", "end"},
		},
		{
			iterator: sliceIterator[string]([]string{"a", "b", "c"}),
			limit:    1,
			expect:   []string{"begin", "accept(a)", "end"},
		},
	} {
		t.Run(fmt.Sprintf("%#v %d", tc.iterator, tc.limit), func(t *testing.T) {
			s := &testSink[string]{limit: tc.limit}

			tc.iterator.copyInto(s)

			if !reflect.DeepEqual(tc.expect, s.log) {
				t.Errorf("wrong result, expected: %v, got: %v", tc.expect, s.log)
			}
		})
	}
}
