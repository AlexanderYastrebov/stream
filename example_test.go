package stream_test

import (
	"fmt"

	"github.com/AlexanderYastrebov/stream"
)

func ExampleCount() {
	n := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Filter(func(s string) bool { return len(s) > 2 }).
		Count()

	fmt.Println(n)

	// Output:
	// 3
}

func ExampleMapReduce() {
	s := stream.Of("a", "bb", "ccc", "dddd")
	n, ok := stream.Map(s, func(s string) int { return len(s) }).
		Reduce(func(a, b int) int { return a + b })

	fmt.Println(n, ok)

	// Output:
	// 10 true
}

func ExampleAllButLast() {
	input := []string{"foo", "bar", "baz", "goo", "bar", "gaz"}
	bars := stream.Slice(input).
		Filter(func(s string) bool { return s == "bar" }).
		Count()
	fmt.Println("bars:", bars)

	result := stream.Slice(input).
		Filter(func(s string) bool {
			if bars > 1 && s == "bar" {
				bars--
				return false
			} else {
				return true
			}
		}).
		ToSlice()

	fmt.Println(result)

	// Output:
	// bars: 2
	// [foo baz goo bar gaz]
}

func ExampleGroupBy() {
	s := stream.Slice([]struct {
		name  string
		grade string
	}{
		{"Alice", "A"},
		{"Bob", "B"},
		{"Charlie", "C"},
		{"Alan", "A"},
		{"Barbie", "B"},
		{"Carl", "C"},
	})

	g := stream.Reduce(s, make(map[string][]string), func(m map[string][]string, e struct {
		name  string
		grade string
	}) map[string][]string {
		m[e.grade] = append(m[e.grade], e.name)
		return m
	})
	fmt.Println(g)

	// Output:
	// map[A:[Alice Alan] B:[Bob Barbie] C:[Charlie Carl]]
}
