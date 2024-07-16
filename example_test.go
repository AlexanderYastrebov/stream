package stream_test

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlexanderYastrebov/stream"
)

func print[T any](x T) {
	fmt.Println(x)
}

func Example() {
	f, _ := os.Open("testdata/lorem.txt")
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)

	stream.While(scanner.Scan, scanner.Text).
		Map(strings.ToLower).
		Map(func(s string) string { return strings.TrimRight(s, ".,") }).
		Filter(stream.Distinct[string]()).
		Sorted(stream.NaturalOrder).
		Limit(10).
		ForEach(print)

	// Output:
	// ad
	// adipiscing
	// aliqua
	// aliquip
	// amet
	// anim
	// aute
	// cillum
	// commodo
	// consectetur
}

func ExampleGenerate() {
	stream.Generate(func() string { return "a" }).
		Limit(3).
		ForEach(print)

	// Output:
	// a
	// a
	// a
}

func ExampleIterate() {
	stream.Iterate(3, func(x int) int { return x + 2 }).
		Limit(4).
		ForEach(print)

	// Output:
	// 3
	// 5
	// 7
	// 9
}

func ExampleIterate_fibonacci() {
	pairs := stream.Iterate([]int{0, 1}, func(x []int) []int {
		return []int{x[1], x[0] + x[1]}
	})

	stream.Map(pairs, func(x []int) int { return x[0] }).
		Limit(10).
		ForEach(print)

	// Output:
	// 0
	// 1
	// 1
	// 2
	// 3
	// 5
	// 8
	// 13
	// 21
	// 34
}

func ExampleStream_filter() {
	n := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Filter(func(s string) bool {
			return len(s) > 2
		}).
		Peek(print).
		Count()

	fmt.Println(n)

	// Output:
	// ccc
	// dddd
	// eeeee
	// 3
}

func ExampleStream_map() {
	stream.Of("a", "bb", "ccc").
		Map(func(s string) string {
			return s + s
		}).
		ForEach(print)

	// Output:
	// aa
	// bbbb
	// cccccc
}

func ExampleStream_skipLimit() {
	n := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Skip(1).
		Limit(3).
		Peek(print).
		Count()

	fmt.Println(n)

	n = stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Limit(0).
		Peek(print).
		Count()

	fmt.Println(n)

	// Output:
	// bb
	// ccc
	// dddd
	// 3
	// 0
}

func ExampleStream_sorted() {
	stream.Of("bb", "a", "dddd", "ccc").
		Sorted(stream.NaturalOrder[string]).
		ForEach(print)

	stream.Of("bb", "a", "dddd", "ccc").
		Sorted(stream.ReverseOrder[string]).
		ForEach(print)

	// Output:
	// a
	// bb
	// ccc
	// dddd
	// dddd
	// ccc
	// bb
	// a
}

func ExampleStream_forEach() {
	stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Filter(func(s string) bool {
			return len(s) > 2
		}).
		ForEach(print)

	// Output:
	// ccc
	// dddd
	// eeeee
}

func ExampleStream_reduce() {
	s := stream.Of("a", "bb", "ccc", "dddd")
	n, ok := stream.Map(s, func(s string) int { return len(s) }).
		Reduce(func(a, b int) int { return a + b })

	fmt.Println(n, ok)

	// Output:
	// 10 true
}

func ExampleFlatMap_sameType() {
	split := func(s string) stream.Stream[string] {
		return stream.Slice(strings.Split(s, ""))
	}

	stream.Of("a", "bb", "ccc", "dddd").
		FlatMap(split).
		Limit(4).
		ForEach(print)

	// Output:
	// a
	// b
	// b
	// c
}

func ExampleFlatMap() {
	runes := func(s string) stream.Stream[rune] {
		return stream.Slice([]rune(s))
	}

	s := stream.Of("a", "bb", "ccc", "dddd")
	stream.FlatMap(s, runes).
		Limit(4).
		ForEach(print)

	// Output:
	// 97
	// 98
	// 98
	// 99
}

func ExampleDistinct() {
	stream.Of("a", "bb", "bb", "a", "ccc", "a", "dddd").
		Filter(stream.Distinct[string]()).
		ForEach(print)

	// Output:
	// a
	// bb
	// ccc
	// dddd
}

func ExampleDistinctUsing() {
	concat := func(s []string) string {
		return strings.Join(s, " ")
	}

	stream.Of(
		[]string{"a"},
		[]string{"b", "b"},
		[]string{"b", "b"},
		[]string{"a"},
		[]string{"c", "c", "c"},
		[]string{"a"},
		[]string{"d", "d", "d", "d"}).
		Filter(stream.DistinctUsing(concat)).
		ForEach(print)

	// Output:
	// [a]
	// [b b]
	// [c c c]
	// [d d d d]
}

func ExampleStream_filterAllButLast() {
	input := []string{"foo", "bar", "baz", "bar", "goo", "bar", "gaz"}
	bars := stream.Slice(input).
		Filter(func(s string) bool {
			return s == "bar"
		}).
		Count()
	fmt.Println("bars:", bars)

	result := stream.Slice(input).
		Filter(func(s string) bool {
			if s == "bar" {
				bars--
				return bars == 0
			}
			return true
		}).
		ToSlice()

	fmt.Println(result)

	// Output:
	// bars: 3
	// [foo baz goo bar gaz]
}

func ExampleCollect_groupBy() {
	type result struct {
		name  string
		grade string
	}

	s := stream.Slice([]result{
		{"Alice", "A"},
		{"Bob", "B"},
		{"Charlie", "C"},
		{"Alan", "A"},
		{"Barbie", "B"},
		{"Carl", "C"},
	})

	g := stream.Collect(s, make(map[string][]string), func(m map[string][]string, r result) {
		m[r.grade] = append(m[r.grade], r.name)
	})
	fmt.Println(g)

	// Output:
	// map[A:[Alice Alan] B:[Bob Barbie] C:[Charlie Carl]]
}

func ExampleStream_allMatch() {
	m := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		AllMatch(func(s string) bool {
			return len(s) > 3
		})

	fmt.Println(m)

	m = stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		AllMatch(func(s string) bool {
			return len(s) > 0
		})

	fmt.Println(m)

	// Output:
	// false
	// true
}

func ExampleStream_anyMatch() {
	m := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		AnyMatch(func(s string) bool {
			return len(s) > 3
		})

	fmt.Println(m)

	m = stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		AnyMatch(func(s string) bool {
			return len(s) > 10
		})

	fmt.Println(m)

	// Output:
	// true
	// false
}

func ExampleStream_noneMatch() {
	m := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		NoneMatch(func(s string) bool {
			return len(s) > 3
		})

	fmt.Println(m)

	m = stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		NoneMatch(func(s string) bool {
			return len(s) > 10
		})

	fmt.Println(m)

	// Output:
	// false
	// true
}

func ExampleStream_findFirst() {
	x, ok := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Filter(func(s string) bool {
			return len(s) > 2
		}).
		FindFirst()

	fmt.Println(x, ok)

	_, ok = stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Filter(func(s string) bool {
			return len(s) > 10
		}).
		FindFirst()

	fmt.Println(ok)

	// Output:
	// ccc true
	// false
}

func ExampleStream_minMax() {
	x, ok := stream.Of(2, 5, 1, 4, 3).
		Max(stream.NaturalOrder[int])

	fmt.Println(x, ok)

	x, ok = stream.Of(2, 5, 1, 4, 3).
		Min(stream.NaturalOrder[int])

	fmt.Println(x, ok)

	x, ok = stream.Of[int]().
		Min(stream.NaturalOrder[int])

	fmt.Println(x, ok)

	// Output:
	// 5 true
	// 1 true
	// 0 false
}

func ExampleStream_append() {
	stream.Of(1, 2, 3).
		Append(stream.Of(4, 5, 6)).
		Filter(func(x int) bool {
			return x%2 == 0
		}).
		ForEach(print)

	stream.Of("a", "bb").
		Append(stream.Of("ccc").Peek(func(string) { panic("oops") })).
		Limit(2). // appended stream is not evaluated
		ForEach(print)

	// Output:
	// 2
	// 4
	// 6
	// a
	// bb
}
