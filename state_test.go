package hyperloglog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		s      state
		expect state
	}{
		{name: "sparse state", s: newState(true), expect: stateSparse},
		{name: "dense state", s: newState(false), expect: stateDense},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expect, tc.s, "Must match the expected state")
		})
	}
}

func TestStateIsSparse(t *testing.T) {
	t.Parallel()

	assert.True(t, stateSparse.isSparse(), "Must be sparse")
	assert.False(t, stateDense.isSparse(), "Must not be sparse")
}

func TestStateByte(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		s    state
		data byte
	}{
		{name: "sparse state", s: newState(true), data: 0x01},
		{name: "dense state", s: newState(false), data: 0x00},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.data, tc.s.Byte(), "Must match the expected byte value")
		})
	}
}
