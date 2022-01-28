package stream

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIterators(t *testing.T) {
	for _, tc := range []struct {
		iterator iterator[string]
		create   func() iterator[string]
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
		{
			iterator: generatorIterator[string](func() string { return "a" }),
			limit:    0,
			expect:   []string{"begin", "end"},
		},
		{
			iterator: generatorIterator[string](func() string { return "a" }),
			limit:    3,
			expect:   []string{"begin", "accept(a)", "accept(a)", "accept(a)", "end"},
		},
		{
			iterator: &seedIterator[string]{"a", func(s string) string { return s + "b" }},
			limit:    0,
			expect:   []string{"begin", "end"},
		},
		{
			iterator: &seedIterator[string]{"a", func(s string) string { return s + "b" }},
			limit:    3,
			expect:   []string{"begin", "accept(a)", "accept(ab)", "accept(abb)", "end"},
		},
		{
			iterator: &whileIterator[string]{func() bool { return false }, func() string { return "a" }},
			limit:    -1,
			expect:   []string{"begin", "end"},
		},
		{
			create: func() iterator[string] {
				counter := 3
				return &whileIterator[string]{
					func() bool {
						return counter > 0
					}, func() (s string) {
						s = fmt.Sprintf("%d", counter)
						counter--
						return
					},
				}
			},
			limit:  -1,
			expect: []string{"begin", "accept(3)", "accept(2)", "accept(1)", "end"},
		},
		{
			create: func() iterator[string] {
				counter := 3
				return &whileIterator[string]{
					func() bool {
						return counter > 0
					}, func() (s string) {
						s = fmt.Sprintf("%d", counter)
						counter--
						return
					},
				}
			},
			limit:  2,
			expect: []string{"begin", "accept(3)", "accept(2)", "end"},
		},
	} {
		t.Run(fmt.Sprintf("%#v %d", tc.iterator, tc.limit), func(t *testing.T) {
			it := tc.iterator
			if tc.create != nil {
				it = tc.create()
			}

			s := &testSink[string]{limit: tc.limit}

			it.copyInto(s)

			if !reflect.DeepEqual(tc.expect, s.log) {
				t.Errorf("wrong result, expected: %v, got: %v", tc.expect, s.log)
			}
		})
	}
}
