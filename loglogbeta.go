package hyperloglog

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
)

// SketchLLB is a HyperLogLog data-structure for the count-distinct problem,
// approximating the number of distinct elements in a multiset.
type SketchLLB struct {
	p          uint8
	m          uint32
	alpha      float64
	tmpSet     set
	sparseList *compressedList
	regs       []uint8
}

// New returns a HyperLogLog Sketch with 2^14 registers (precision 14)
func NewLLB() *SketchLLB {
	return NewLLB14()
}

// New14 returns a HyperLogLog Sketch with 2^14 registers (precision 14)
func NewLLB14() *SketchLLB {
	sk, _ := newSketchLLB(14, true)
	return sk
}

// New16 returns a HyperLogLog Sketch with 2^16 registers (precision 16)
func NewLLB16() *SketchLLB {
	sk, _ := newSketchLLB(16, true)
	return sk
}

// NewNoSparse returns a HyperLogLog Sketch with 2^14 registers (precision 14)
// that will not use a sparse representation
func NewLLBNoSparse() *SketchLLB {
	sk, _ := newSketchLLB(14, false)
	return sk
}

// New16NoSparse returns a HyperLogLog Sketch with 2^16 registers (precision 16)
// that will not use a sparse representation
func NewLLB16NoSparse() *SketchLLB {
	sk, _ := newSketchLLB(16, false)
	return sk
}

// newSketch returns a HyperLogLog Sketch with 2^precision registers
func newSketchLLB(precision uint8, sparse bool) (*SketchLLB, error) {
	if precision < 4 || precision > 18 {
		return nil, fmt.Errorf("p has to be >= 4 and <= 18")
	}
	m := uint32(math.Pow(2, float64(precision)))
	s := &SketchLLB{
		m:     m,
		p:     precision,
		alpha: alpha(float64(m)),
	}
	if sparse {
		s.tmpSet = set{}
		s.sparseList = newCompressedList(0)
	} else {
		s.regs = make([]uint8, m)
	}
	return s, nil
}

func (sk *SketchLLB) sparse() bool {
	return sk.sparseList != nil
}

// Clone returns a deep copy of sk.
func (sk *SketchLLB) Clone() *SketchLLB {
	regs := make([]uint8, len(sk.regs))
	copy(regs, sk.regs)
	return &SketchLLB{
		p:          sk.p,
		m:          sk.m,
		alpha:      sk.alpha,
		tmpSet:     sk.tmpSet.Clone(),
		sparseList: sk.sparseList.Clone(),
		regs:       regs,
	}
}

// Converts to normal if the sparse list is too large.
func (sk *SketchLLB) maybeToNormal() {
	if uint32(len(sk.tmpSet))*100 > sk.m {
		sk.mergeSparse()
		if uint32(sk.sparseList.Len()) > sk.m {
			sk.toNormal()
		}
	}
}

// Merge takes another Sketch and combines it with Sketch h.
// If Sketch h is using the sparse Sketch, it will be converted
// to the normal Sketch.
func (sk *SketchLLB) Merge(other *SketchLLB) error {
	if other == nil {
		// Nothing to do
		return nil
	}
	cpOther := other.Clone()

	if sk.p != cpOther.p {
		return errors.New("precisions must be equal")
	}

	if sk.sparse() && other.sparse() {
		for k := range other.tmpSet {
			sk.tmpSet.add(k)
		}
		for iter := other.sparseList.Iter(); iter.HasNext(); {
			sk.tmpSet.add(iter.Next())
		}
		sk.maybeToNormal()
		return nil
	}

	if sk.sparse() {
		sk.toNormal()
	}

	if cpOther.sparse() {
		for k := range cpOther.tmpSet {
			i, r := decodeHash(k, cpOther.p, pp)
			sk.insert(i, r)
		}

		for iter := cpOther.sparseList.Iter(); iter.HasNext(); {
			i, r := decodeHash(iter.Next(), cpOther.p, pp)
			sk.insert(i, r)
		}
	} else {
		for i, v := range cpOther.regs {
			if v > sk.regs[i] {
				sk.regs[i] = v
			}
		}
	}
	return nil
}

// Convert from sparse Sketch to dense Sketch.
func (sk *SketchLLB) toNormal() {
	if len(sk.tmpSet) > 0 {
		sk.mergeSparse()
	}

	sk.regs = make([]uint8, sk.m)
	for iter := sk.sparseList.Iter(); iter.HasNext(); {
		i, r := decodeHash(iter.Next(), sk.p, pp)
		sk.insert(i, r)
	}

	sk.tmpSet = nil
	sk.sparseList = nil
}

func (sk *SketchLLB) insert(i uint32, r uint8) bool {
	changed := false
	if r > sk.regs[i] {
		sk.regs[i] = r
		changed = true
	}
	return changed
}

// Insert adds element e to sketch
func (sk *SketchLLB) Insert(e []byte) bool {
	x := hash(e)
	return sk.InsertHash(x)
}

// InsertHash adds hash x to sketch
func (sk *SketchLLB) InsertHash(x uint64) bool {
	if sk.sparse() {
		changed := sk.tmpSet.add(encodeHash(x, sk.p, pp))
		if !changed {
			return false
		}
		if uint32(len(sk.tmpSet))*100 > sk.m {
			sk.mergeSparse()
			if uint32(sk.sparseList.Len()) > sk.m {
				sk.toNormal()
			}
		}
		return true
	} else {
		i, r := getPosVal(x, sk.p)
		return sk.insert(uint32(i), r)
	}
}

// Estimate returns the cardinality of the Sketch
func (sk *SketchLLB) Estimate() uint64 {
	if sk.sparse() {
		sk.mergeSparse()
		return uint64(linearCount(mp, mp-sk.sparseList.count))
	}

	sum, ez := sumAndZeros(sk.regs)
	m := float64(sk.m)
	var est float64

	var beta func(float64) float64
	if sk.p < 16 {
		beta = beta14
	} else {
		beta = beta16
	}

	est = sk.alpha * m * (m - ez) / (sum + beta(ez))
	return uint64(est + 0.5)
}

func (sk *SketchLLB) mergeSparse() {
	if len(sk.tmpSet) == 0 {
		return
	}

	keys := make(uint64Slice, 0, len(sk.tmpSet))
	for k := range sk.tmpSet {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	newList := newCompressedList(4*len(sk.tmpSet) + len(sk.sparseList.b))
	for iter, i := sk.sparseList.Iter(), 0; iter.HasNext() || i < len(keys); {
		if !iter.HasNext() {
			newList.Append(keys[i])
			i++
			continue
		}

		if i >= len(keys) {
			newList.Append(iter.Next())
			continue
		}

		x1, x2 := iter.Peek(), keys[i]
		if x1 == x2 {
			newList.Append(iter.Next())
			i++
		} else if x1 > x2 {
			newList.Append(x2)
			i++
		} else {
			newList.Append(iter.Next())
		}
	}

	sk.sparseList = newList
	sk.tmpSet = set{}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (sk *SketchLLB) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 0, len(sk.regs)+8)

	// Marshal a version marker.
	data = append(data, version)
	// Marshal p.
	data = append(data, sk.p)
	// Marshal b
	data = append(data, 0)

	if sk.sparse() {
		// It's using the sparse Sketch.
		data = append(data, byte(1))

		// Add the tmp_set
		tsdata, err := sk.tmpSet.MarshalBinary()
		if err != nil {
			return nil, err
		}
		data = append(data, tsdata...)

		// Add the sparse Sketch
		sdata, err := sk.sparseList.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(data, sdata...), nil
	}

	// It's using the dense Sketch.
	data = append(data, byte(0))

	// Add the dense sketch Sketch.
	sz := len(sk.regs)
	data = append(data, []byte{
		byte(sz >> 24),
		byte(sz >> 16),
		byte(sz >> 8),
		byte(sz),
	}...)

	// Marshal each element in the list.
	for _, v := range sk.regs {
		data = append(data, byte(v))
	}

	return data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (sk *SketchLLB) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return ErrorTooShort
	}

	// Unmarshal version. We may need this in the future if we make
	// non-compatible changes.
	v := data[0]

	// Unmarshal p.
	p := data[1]

	// Unmarshal b.
	b := data[2]

	// Determine if we need a sparse Sketch
	sparse := data[3] == byte(1)

	// Make a newSketch Sketch if the precision doesn't match or if the Sketch was used
	if sk.p != p || sk.regs != nil || len(sk.tmpSet) > 0 || (sk.sparseList != nil && sk.sparseList.Len() > 0) {
		newh, err := newSketchLLB(p, sparse)
		if err != nil {
			return err
		}
		*sk = *newh
	}

	// h is now initialised with the correct p. We just need to fill the
	// rest of the details out.
	if sparse {
		// Using the sparse Sketch.

		// Unmarshal the tmp_set.
		tssz := binary.BigEndian.Uint32(data[4:8])
		sk.tmpSet = make(map[uint32]struct{}, tssz)

		// We need to unmarshal tssz values in total, and each value requires us
		// to read 4 bytes.
		tsLastByte := int((tssz * 4) + 8)
		for i := 8; i < tsLastByte; i += 4 {
			k := binary.BigEndian.Uint32(data[i : i+4])
			sk.tmpSet[k] = struct{}{}
		}

		// Unmarshal the sparse Sketch.
		return sk.sparseList.UnmarshalBinary(data[tsLastByte:])
	}

	// Using the dense Sketch.
	sk.sparseList = nil
	sk.tmpSet = nil

	if v == 1 {
		return sk.unmarshalBinaryV1(data[8:], b)
	}
	return sk.unmarshalBinaryV2(data)
}

func sumAndZeros(regs []uint8) (res, ez float64) {
	for _, v := range regs {
		if v == 0 {
			ez++
		}
		res += 1.0 / math.Pow(2.0, float64(v))
	}
	return res, ez
}

func (sk *SketchLLB) unmarshalBinaryV1(data []byte, b uint8) error {
	sk.regs = make([]uint8, len(data)*2)
	for i, v := range data {
		sk.regs[i*2] = uint8((v >> 4)) + b
		sk.regs[i*2+1] = uint8((v<<4)>>4) + b
	}
	return nil
}

func (sk *SketchLLB) unmarshalBinaryV2(data []byte) error {
	sk.regs = data[8:]
	return nil
}
