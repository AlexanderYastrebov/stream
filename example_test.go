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
