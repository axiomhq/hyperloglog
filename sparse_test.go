package hyperloglog

import (
	"math/rand"
	"testing"
)

func TestSparseEncodeDecode(t *testing.T) {
	const p, pp = 14, 25
	for i := 0; i < 1000000; i++ {
		x := rand.Uint64()
		k := encodeHash(x, p, pp)
		idx1, rho1 := decodeHash(k, p, pp)
		idx2, rho2 := getPosVal(x, p)
		if uint64(idx1) != idx2 || rho1 != rho2 {
			t.Fatalf("decode failure: i=%d x=%016x k=%08x idx1=%08x rho1=%02x idx2=%08x rho2=%02x", i, x, k, idx1, rho1, idx2, rho2)
		}
	}
}
