package hlltc

import (
	"fmt"
	"testing"

	"github.com/clarkduvall/hyperloglog"
	metro "github.com/dgryski/go-metro"
)

type fakeHash64 uint64

func (f fakeHash64) Sum64() uint64 { return uint64(f) }

func estimateError(got, exp uint64) float64 {
	var delta uint64
	if got > exp {
		delta = got - exp
	} else {
		delta = exp - got
	}
	return float64(delta) / float64(exp)
}

func TestCardinality(t *testing.T) {
	hlltc, err := New(14)
	hllpp, err := hyperloglog.NewPlus(14)
	if err != nil {
		t.Error("expected no error, got", err)
	}

	step := 10
	unique := map[string]bool{}

	for i := 1; len(unique) <= 10000000; i++ {
		str := fmt.Sprintf("flow-%d", i)
		hlltc.Insert([]byte(str))
		item := fakeHash64(metro.Hash64([]byte(str), 1337))
		hllpp.Add(item)
		unique[str] = true

		if len(unique)%step == 0 {
			step *= 5
			exact := uint64(len(unique))
			res := uint64(hlltc.Estimate())
			ratio := 100 * estimateError(res, exact)
			if ratio > 2 {
				t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
			}
			res2 := uint64(hllpp.Count())
			ratio2 := 100 * estimateError(res2, exact)
			fmt.Printf("Exact %d, got:\n\thlltc %d (%.2f%% off)\n\thllpp %d (%.2f%% off)\n", exact, res, ratio, res2, ratio2)
		}
	}
	exact := uint64(len(unique))
	res := uint64(hlltc.Estimate())
	ratio := 100 * estimateError(res, exact)
	if ratio > 2 {
		t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
	}
	res2 := uint64(hllpp.Count())
	ratio2 := 100 * estimateError(res2, exact)
	fmt.Printf("Exact %d, got:\n\thlltc %d (%.2f%% off)\n\thllpp %d (%.2f%% off)\n", exact, res, ratio, res2, ratio2)
}
