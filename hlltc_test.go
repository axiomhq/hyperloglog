package hlltc

import (
	"fmt"
	"testing"
)

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
	if err != nil {
		t.Error("expected no error, got", err)
	}

	step := 10
	unique := map[string]bool{}

	for i := 1; len(unique) <= 10000000; i++ {
		str := fmt.Sprintf("flow-%d", i)
		hlltc.Insert([]byte(str))
		unique[str] = true

		if len(unique)%step == 0 {
			step *= 5
			exact := uint64(len(unique))
			res := uint64(hlltc.Estimate())
			ratio := 100 * estimateError(res, exact)
			if ratio > 2 {
				t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
			}
		}
	}
	exact := uint64(len(unique))
	res := uint64(hlltc.Estimate())
	ratio := 100 * estimateError(res, exact)
	if ratio > 2 {
		t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
	}
}
