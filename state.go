package hyperloglog

// state allows for ease of tracking the current representation
// of the sketch to allow for ease of restoring sketch state and reuse.
type state byte

const (
	stateDense state = iota
	stateSparse
)

func newState(sparse bool) state {
	if sparse {
		return stateSparse
	}
	return stateDense
}

func (s state) isSparse() bool {
	return s == stateSparse
}

func (s state) Byte() byte {
	return byte(s)
}
