package hyperloglog

import (
	"encoding/binary"
	"math"
	"math/bits"
)

var hash = hashFunc

func beta14(ez float64) float64 {
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

func beta16(ez float64) float64 {
	zl := math.Log(ez + 1)
	return -0.37331876643753059*ez +
		-1.41704077448122989*zl +
		0.40729184796612533*math.Pow(zl, 2) +
		1.56152033906584164*math.Pow(zl, 3) +
		-0.99242233534286128*math.Pow(zl, 4) +
		0.26064681399483092*math.Pow(zl, 5) +
		-0.03053811369682807*math.Pow(zl, 6) +
		0.00155770210179105*math.Pow(zl, 7)
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
	rho := uint8(bits.LeadingZeros64(w)) + 1
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

func hashFunc(e []byte) uint64 {
	return hash64(e, 1337)
}

func rotateRight(v uint64, k uint) uint64 {
	return (v >> k) | (v << (64 - k))
}

func hash64(buffer []byte, seed uint64) uint64 {

	const (
		k0 = 0xD6D018F5
		k1 = 0xA2AA033B
		k2 = 0x62992FC1
		k3 = 0x30BC5B29
	)

	ptr := buffer

	hash := (seed + k2) * k0

	if len(ptr) >= 32 {
		v := [4]uint64{hash, hash, hash, hash}

		for len(ptr) >= 32 {
			v[0] += binary.LittleEndian.Uint64(ptr[:8]) * k0
			v[0] = rotateRight(v[0], 29) + v[2]
			v[1] += binary.LittleEndian.Uint64(ptr[8:16]) * k1
			v[1] = rotateRight(v[1], 29) + v[3]
			v[2] += binary.LittleEndian.Uint64(ptr[16:24]) * k2
			v[2] = rotateRight(v[2], 29) + v[0]
			v[3] += binary.LittleEndian.Uint64(ptr[24:32]) * k3
			v[3] = rotateRight(v[3], 29) + v[1]
			ptr = ptr[32:]
		}

		v[2] ^= rotateRight(((v[0]+v[3])*k0)+v[1], 37) * k1
		v[3] ^= rotateRight(((v[1]+v[2])*k1)+v[0], 37) * k0
		v[0] ^= rotateRight(((v[0]+v[2])*k0)+v[3], 37) * k1
		v[1] ^= rotateRight(((v[1]+v[3])*k1)+v[2], 37) * k0
		hash += v[0] ^ v[1]
	}

	if len(ptr) >= 16 {
		v0 := hash + (binary.LittleEndian.Uint64(ptr[:8]) * k2)
		v0 = rotateRight(v0, 29) * k3
		v1 := hash + (binary.LittleEndian.Uint64(ptr[8:16]) * k2)
		v1 = rotateRight(v1, 29) * k3
		v0 ^= rotateRight(v0*k0, 21) + v1
		v1 ^= rotateRight(v1*k3, 21) + v0
		hash += v1
		ptr = ptr[16:]
	}

	if len(ptr) >= 8 {
		hash += binary.LittleEndian.Uint64(ptr[:8]) * k3
		ptr = ptr[8:]
		hash ^= rotateRight(hash, 55) * k1
	}

	if len(ptr) >= 4 {
		hash += uint64(binary.LittleEndian.Uint32(ptr[:4])) * k3
		hash ^= rotateRight(hash, 26) * k1
		ptr = ptr[4:]
	}

	if len(ptr) >= 2 {
		hash += uint64(binary.LittleEndian.Uint16(ptr[:2])) * k3
		ptr = ptr[2:]
		hash ^= rotateRight(hash, 48) * k1
	}

	if len(ptr) >= 1 {
		hash += uint64(ptr[0]) * k3
		hash ^= rotateRight(hash, 37) * k1
	}

	hash ^= rotateRight(hash, 28)
	hash *= k0
	hash ^= rotateRight(hash, 29)

	return hash
}
