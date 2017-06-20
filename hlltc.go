package hlltc

import (
	"fmt"
	"math"
	"sort"

	metro "github.com/dgryski/go-metro"
)

const (
	capacity = uint8(16)
	pp       = uint8(25)
	mp       = uint32(1) << pp
)

// Sketch ...
type Sketch struct {
	regs       *rs
	m          uint32
	precision  uint8
	b          uint8
	alpha      float64
	sparse     bool
	sparseList *compressedList
	tmpSet     set
}

// New ...
func New(precision uint8) (*Sketch, error) {
	if precision < 6 || precision > 16 {
		return nil, fmt.Errorf("precision has to be >= 8 and <= 16")
	}
	m := uint32(math.Pow(2, float64(precision)))
	return &Sketch{
		m:          m,
		precision:  precision,
		alpha:      alpha(float64(m)),
		sparse:     true,
		tmpSet:     set{},
		sparseList: newCompressedList(int(m)),
	}, nil
}

// Convert from sparse representation to dense representation.
func (sk *Sketch) toNormal() {
	if len(sk.tmpSet) > 0 {
		sk.mergeSparse()
	}

	sk.regs = newRegs(sk.m)
	for iter := sk.sparseList.Iter(); iter.HasNext(); {
		i, r := decodeHash(iter.Next(), sk.precision, pp)
		sk.insert(i, r)
	}

	sk.sparse = false
	sk.tmpSet = nil
	sk.sparseList = nil
}

func (sk *Sketch) insert(i uint32, r uint8) {
	if r-sk.b >= capacity {
		//overflow
		db := sk.regs.min()
		if db > 0 {
			sk.b += db
			sk.regs.rebase(db)
		}
	}
	if r > sk.b {
		val := uint8(math.Min(float64(r-sk.b), float64(capacity-1)))
		if val > sk.regs.get(i) {
			sk.regs.set(i, uint8(val))
		}
	}
}

// Insert ...
func (sk *Sketch) Insert(e []byte) {
	x := metro.Hash64(e, 42)
	if sk.sparse {
		sk.tmpSet.add(encodeHash(x, sk.precision, pp))
		if uint32(len(sk.tmpSet))*100 > sk.m {
			sk.mergeSparse()
			if uint32(sk.sparseList.Len()) > sk.m {
				sk.toNormal()
			}
		}
	} else {
		i, r := getPosVal(x, sk.precision)
		sk.insert(uint32(i), r)
	}
}

// Estimate ...
func (sk *Sketch) Estimate() uint64 {
	if sk.sparse {
		sk.mergeSparse()
		return uint64(linearCount(mp, mp-uint32(sk.sparseList.count)))
	}

	sum := float64(sk.regs.sum(sk.b))
	ez := float64(sk.regs.zeros())
	m := float64(sk.m)
	est := sk.alpha * m * m / sum

	if sk.b == 0 {
		return uint64((sk.alpha * m * (m - ez) / (sum + beta(ez))) + 0.5)
	}
	return uint64(est + 0.5)

}

func (sk *Sketch) mergeSparse() {
	if len(sk.tmpSet) == 0 {
		return
	}

	keys := make(uint64Slice, 0, len(sk.tmpSet))
	for k := range sk.tmpSet {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	newList := newCompressedList(int(sk.m))
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
