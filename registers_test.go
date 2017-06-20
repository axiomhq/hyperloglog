package hlltc

import (
	"math/rand"
	"testing"
)

func TestRegistersGetSetSum(t *testing.T) {
	length := uint32(16777216)
	data := make([]uint8, length, length)
	r := newRegs(length)

	for i := range data {
		val := uint8(rand.Intn(16))
		r.set(uint32(i), val)
		data[i] = val
	}

	for i, exp := range data {
		if got := r.get(uint32(i)); exp != got {
			t.Errorf("expected %d, got %d", exp, got)
		}
	}
}

func TestRegistersZeros(t *testing.T) {
	m := uint32(2)
	rs := newRegs(m)
	for i := uint32(0); i < m; i++ {
		rs.set(i, uint8(i)%15)
	}
	for i := uint32(0); i < m; i++ {
		exp := uint8(i % 15)
		if got := rs.get(i); got != exp {
			t.Errorf("expected %d, got %d", exp, got)
		}
	}

}
