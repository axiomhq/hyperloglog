package hlltc

import (
	"math"

	bits "github.com/dgryski/go-bits"
	metro "github.com/dgryski/go-metro"
)

func beta(ez float64) float64 {
	zl := math.Log(ez + 1)
	return -0.370393911*ez +
		0.070471823*zl +
		0.17393686*math.Pow(zl, 2) +
		0.16339839*math.Pow(zl, 3) +
		-0.09237745*math.Pow(zl, 4) +
		0.03738027*math.Pow(zl, 5) +
		-0.005384159*math.Pow(zl, 6) +
		0.00042419*math.Pow(zl, 7)
}

func alpha(m float64) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	}
	return 0.7213 / (1 + 1.079/m)
}

func getPosVal(x uint64, p uint8) (uint64, uint8) {
	i := bextr(x, 64-p, p) // {x63,...,x64-p}
	w := x<<p | 1<<(p-1)   // {x63-p,...,x0}
	rho := uint8(bits.Clz(w)) + 1
	return i, rho
}

func linearCount(m uint32, v uint32) float64 {
	fm := float64(m)
	return fm * math.Log(fm/float64(v))
}

func bextr(v uint64, start, length uint8) uint64 {
	return (v >> start) & ((1 << length) - 1)
}

func bextr32(v uint32, start, length uint8) uint32 {
	return (v >> start) & ((1 << length) - 1)
}

func getIndex(k uint32, p, pp uint8) uint32 {
	if k&1 == 1 {
		return bextr32(k, 32-p, p)
	}
	return bextr32(k, pp-p+1, p)
}

// Encode a hash to be used in the sparse representation.
func encodeHash(x uint64, p, pp uint8) uint32 {
	idx := uint32(bextr(x, 64-pp, pp))
	if bextr(x, 64-pp, pp-p) == 0 {
		zeros := bits.Clz((bextr(x, 0, 64-pp)<<pp)|(1<<pp-1)) + 1
		return idx<<7 | uint32(zeros<<1) | 1
	}
	return idx << 1
}

// Decode a hash from the sparse representation.
func decodeHash(k uint32, p, pp uint8) (uint32, uint8) {
	var r uint8
	if k&1 == 1 {
		r = uint8(bextr32(k, 1, 6)) + pp - p
	} else {
		// We can use the 64bit clz implementation and reduce the result
		// by 32 to get a clz for a 32bit word.
		r = uint8(bits.Clz(uint64(k<<(32-pp+p-1))) - 31) // -32 + 1
	}
	return getIndex(k, p, pp), r
}

type set map[uint32]struct{}

func (s set) add(v uint32)      { s[v] = struct{}{} }
func (s set) has(v uint32) bool { _, ok := s[v]; return ok }

func (s set) MarshalBinary() (data []byte, err error) {
	// 4 bytes for the size of the set, and 4 bytes for each key.
	// list.
	data = make([]byte, 0, 4+(4*len(s)))

	// Length of the set. We only need 32 bits because the size of the set
	// couldn't exceed that on 32 bit architectures.
	sl := len(s)
	data = append(data, []byte{
		byte(sl >> 24),
		byte(sl >> 16),
		byte(sl >> 8),
		byte(sl),
	}...)

	// Marshal each element in the set.
	for k := range s {
		data = append(data, []byte{
			byte(k >> 24),
			byte(k >> 16),
			byte(k >> 8),
			byte(k),
		}...)
	}

	return data, nil
}

type uint64Slice []uint32

func (p uint64Slice) Len() int           { return len(p) }
func (p uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func hash(e []byte) uint64 {
	return metro.Hash64(e, 1337)
}
