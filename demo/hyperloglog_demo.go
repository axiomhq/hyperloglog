package main

import (
	"fmt"
	"strconv"

	hyperloglog "github.com/influxdata/influxdb/pkg/estimator/hll"
	"github.com/seiflotfy/hlltc"
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

func main() {
	hlltc, _ := hlltc.New(14)
	hllpp, _ := hyperloglog.NewPlus(14)

	step := 10
	unique := map[string]bool{}

	for i := 1; len(unique) <= 10000000; i++ {
		str := "stream-" + strconv.Itoa(i)
		hlltc.Insert([]byte(str))
		hllpp.Add([]byte(str))
		unique[str] = true

		if len(unique)%step == 0 || len(unique) == 10000000 {
			step *= 5
			exact := uint64(len(unique))
			res := uint64(hlltc.Estimate())
			ratio := 100 * estimateError(res, exact)
			res2 := uint64(hllpp.Count())
			ratio2 := 100 * estimateError(res2, exact)
			fmt.Printf("Exact %d, got:\n\t axiom hlltc %d (%.4f%% off)\n\tinflux hllpp %d (%.4f%% off)\n", exact, res, ratio, res2, ratio2)
		}
	}

	data1, _ := hllpp.MarshalBinary()
	data2, _ := hlltc.MarshalBinary()
	fmt.Println("AxiomHQ HLLTC total size", len(data2))
	fmt.Println("InfluxData HLLPP total size", len(data1))

}
