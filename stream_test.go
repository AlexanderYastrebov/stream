package stream

import "testing"

func TestPipeline(t *testing.T) {
	q := Of("_").
		Filter(func(s string) bool { return true })

	r := Map(q, func(s string) int { return len(s) }).
		Filter(func(i int) bool { return true })

	u := Map(r, func(i int) float64 { return float64(i) })

	s := &chainedSink[float64, float64]{
		downstream: nil,
		acceptFunc: func(x float64) {
			t.Logf("accept: %f", x)
		},
	}

	rr := u.(*pipeline[string, float64])

	var x sink[string]

	rr.wrapSink(s, func(_ iterator[string], s sink[string]) {
		x = s
	})

	x.accept("test1")
	x.accept("test22")
}
