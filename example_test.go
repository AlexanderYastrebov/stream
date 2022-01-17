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
