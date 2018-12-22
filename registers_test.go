package hyperloglog

import (
	"math/rand"
	"testing"
)

func TestRegistersGetSetSum(t *testing.T) {
	length := uint32(16777216)
	data := make([]uint8, length)
	r := newRegisters(length)

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
	m := uint32(8)
	rs := newRegisters(m)
	for i := uint32(0); i < m; i++ {
		rs.set(i, (uint8(i)%15)+1)
	}
	for i := uint32(0); i < m; i++ {
		rs.set(i, (uint8(i)%15)+1)
	}
	for i := uint32(0); i < m; i++ {
		exp := uint8(i%15) + 1
		if got := rs.get(i); got != exp {
			t.Errorf("expected %d, got %d", exp, got)
		}
	}

	rs.rebase(1)

	for i := uint32(0); i < m; i++ {
		exp := uint8(i % 15)
		if got := rs.get(i); got != exp {
			t.Errorf("expected %d, got %d", exp, got)
		}
	}

	if got := rs.nz; got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func assertRegistersNz(t *testing.T, rs *registers, nz uint32) {
	t.Helper()
	if rs.nz != nz {
		t.Fatalf("registers.nz is not %d: actual=%d", nz, rs.nz)
	}
}

func TestRegistersSetRepeatedly(t *testing.T) {
	rs := newRegisters(16)

	// count down nz when set non-zero values to each registers.
	assertRegistersNz(t, rs, 16)
	for i := uint32(0); i < 16; i++ {
		rs.set(i, 1)
		assertRegistersNz(t, rs, 15 - i)
	}

	// keep nz:0 when set non-zero values.
	for i := uint8(1); i <= 15; i++ {
		for j := uint32(0); j < 16; j++ {
			rs.set(j, i)
			if rs.nz != 0 {
				t.Fatalf("registers.nz is not zero: actual=%d (i=%d, j=%d)", rs.nz, i, j)
			}
		}
	}
}
