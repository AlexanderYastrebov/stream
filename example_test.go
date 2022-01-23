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

func ExampleWordsDict() {
	f, _ := os.Open("testdata/lorem.txt")
	defer f.Close()

	scanner := bufio.NewScanner(f)

	lines := stream.While(scanner.Scan, scanner.Text)
	words := stream.FlatMap(lines, func(line string) stream.Stream[string] {
		return stream.Slice(strings.Split(line, " "))
	})
	stream.Map(words, strings.ToLower).
		Filter(stream.Distinct[string]()).
		Sorted(stream.NaturalOrder[string]).
		Limit(10).
		ForEach(print[string])

	// Output:
	// ad
	// adipiscing
	// aliqua.
	// aliquip
	// amet,
	// anim
	// aute
	// cillum
	// commodo
	// consectetur
}

func ExampleGenerate() {
	stream.Generate(func() string { return "a" }).
		Limit(3).
		ForEach(print[string])

	// Output:
	// a
	// a
	// a
}

func ExampleIterate() {
	stream.Iterate(3, func(x int) int { return x + 2 }).
		Limit(4).
		ForEach(print[int])

	// Output:
	// 3
	// 5
	// 7
	// 9
}

func ExampleFib() {
	pairs := stream.Iterate([]int{0, 1}, func(x []int) []int {
		return []int{x[1], x[0] + x[1]}
	})

	stream.Map(pairs, func(x []int) int { return x[0] }).
		Limit(10).
		ForEach(print[int])

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

func ExampleFilter() {
	n := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Filter(func(s string) bool {
			return len(s) > 2
		}).
		Peek(print[string]).
		Count()

	fmt.Println(n)

	// Output:
	// ccc
	// dddd
	// eeeee
	// 3
}

func ExampleSkipLimit() {
	n := stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Skip(1).
		Limit(3).
		Peek(print[string]).
		Count()

	fmt.Println(n)

	n = stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Limit(0).
		Peek(print[string]).
		Count()

	fmt.Println(n)

	// Output:
	// bb
	// ccc
	// dddd
	// 3
	// 0
}

func ExampleSorted() {
	stream.Of("bb", "a", "dddd", "ccc").
		Sorted(stream.NaturalOrder[string]).
		ForEach(print[string])

	stream.Of("bb", "a", "dddd", "ccc").
		Sorted(stream.ReverseOrder[string]).
		ForEach(print[string])

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

func ExampleForEach() {
	stream.Of("a", "bb", "ccc", "dddd", "eeeee").
		Filter(func(s string) bool {
			return len(s) > 2
		}).
		ForEach(print[string])

	// Output:
	// ccc
	// dddd
	// eeeee
}

func ExampleMapReduce() {
	s := stream.Of("a", "bb", "ccc", "dddd")
	n, ok := stream.Map(s, func(s string) int { return len(s) }).
		Reduce(func(a, b int) int { return a + b })

	fmt.Println(n, ok)

	// Output:
	// 10 true
}

func ExampleFlatMap() {
	split := func(s string) stream.Stream[string] {
		return stream.Slice(strings.Split(s, ""))
	}

	s := stream.Of("a", "bb", "ccc", "dddd")
	stream.FlatMap(s, split).
		Limit(4).
		ForEach(print[string])

	// Output:
	// a
	// b
	// b
	// c
}

func ExampleDistinct() {
	stream.Of("a", "bb", "bb", "a", "ccc", "a", "dddd").
		Filter(stream.Distinct[string]()).
		ForEach(print[string])

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
		ForEach(print[[]string])

	// Output:
	// [a]
	// [b b]
	// [c c c]
	// [d d d d]
}

func ExampleAllButLast() {
	input := []string{"foo", "bar", "baz", "goo", "bar", "gaz"}
	bars := stream.Slice(input).
		Filter(func(s string) bool {
			return s == "bar"
		}).
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

func ExampleAllMatch() {
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

func ExampleAnyMatch() {
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

func ExampleNoneMatch() {
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

func ExampleFindFirst() {
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

func ExampleMax() {
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
