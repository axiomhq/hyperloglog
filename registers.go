package hlltc

import (
	"math"
)

type reg uint8
type fields []reg

type registers struct {
	fields
	nz uint32
}

func (r *reg) set(offset, val uint8) bool {
	isZero := false
	if offset == 0 {
		isZero = uint8((*r)>>4) == 0
		tmpVal := uint8((*r) << 4 >> 4)
		*r = reg(uint8(tmpVal) | (val << 4))
	} else {
		isZero = uint8((*r)<<4>>4) == 0
		tmpVal := uint8((*r) >> 4)
		*r = reg(tmpVal<<4 | val)
	}
	return isZero
}

func (r *reg) get(offset uint8) uint8 {
	if offset == 0 {
		return uint8((*r) >> 4)
	}
	return uint8((*r) << 4 >> 4)
}

func newRegisters(size uint32) *registers {
	return &registers{
		fields: make(fields, size/2, size/2),
		nz:     size,
	}
}

func (rs *registers) rebase(delta uint8) {
	nz := uint32(len(rs.fields)) * 2
	for i := range rs.fields {
		val := rs.fields[i].get(0)
		if val >= delta {
			rs.fields[i].set(0, val-delta)
			if val-delta > 0 {
				nz--
			}
		}
		val = rs.fields[i].get(1)
		if val >= delta {
			rs.fields[i].set(1, val-delta)
			if val-delta > 0 {
				nz--
			}
		}
	}
	rs.nz = nz
}

func (rs *registers) set(i uint32, val uint8) {
	offset, index := uint8(i%2), i/2
	if rs.fields[index].set(offset, val) {
		rs.nz--
	}
}

func (rs *registers) get(i uint32) uint8 {
	offset, index := uint8(i%2), i/2
	return rs.fields[index].get(offset)
}

func (rs *registers) sum(base uint8) (res float64) {
	for _, r := range rs.fields {
		res += 1.0 / math.Pow(2.0, float64(base+r.get(0)))
		res += 1.0 / math.Pow(2.0, float64(base+r.get(1)))
	}
	return res
}

func (rs *registers) zeros() (res uint32) {
	return rs.nz
}

func (rs *registers) min() uint8 {
	if rs.nz > 0 {
		return 0
	}
	min := uint8(math.MaxUint8)
	for _, r := range rs.fields {
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
