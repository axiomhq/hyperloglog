package hlltc

import (
	"math"
)

type reg uint8
type regs []reg
type rs struct {
	regs
	nz uint32
}

func newRegs(size uint32) *rs {
	return &rs{
		regs: make(regs, size/2, size/2),
		nz:   size,
	}
}

func (r *reg) set(offset, val uint8) {
	if offset == 0 {
		tmpVal := uint8((*r) << 4 >> 4)
		*r = reg(uint8(tmpVal) | (val << 4))
	} else {
		tmpVal := uint8((*r) >> 4)
		*r = reg(tmpVal<<4 | val)
	}
}

func (r *reg) get(offset uint8) uint8 {
	if offset == 0 {
		return uint8((*r) >> 4)
	}
	return uint8((*r) << 4 >> 4)
}

func (rs *rs) rebase(delta uint8) {
	db := delta<<4 | delta
	for i, r := range rs.regs {
		rs.regs[i] = r - reg(db)
	}
}

func (rs *rs) set(i uint32, val uint8) {
	offset, index := uint8(i%2), i/2
	rs.regs[index].set(offset, val)
}

func (rs *rs) get(i uint32) uint8 {
	offset, index := uint8(i%2), i/2
	return rs.regs[index].get(offset)
}

func (rs *rs) sum(base uint8) (res float64) {
	for _, r := range rs.regs {
		res += 1.0 / math.Pow(2.0, float64(base+r.get(0)))
		res += 1.0 / math.Pow(2.0, float64(base+r.get(1)))
	}
	return res
}

func (rs *rs) zeros() (res uint64) {
	for _, r := range rs.regs {
		if uint8(r<<4) == 0 {
			res++
		}
		if r < 16 {
			res++
		}
	}
	return res
}

func (rs *rs) min() uint8 {
	min := uint8(math.MaxUint8)
	for _, r := range rs.regs {
		if val := uint8(r << 4 >> 4); val < min {
			min = val
		}
		if val := uint8(r >> 4); val < min {
			min = val
		}
		if min == 0 {
			break
		}
	}
	return min
}
