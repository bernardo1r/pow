package rand

import (
	cryptoRand "crypto/rand"
)

const buffSize = 20 * (1 << 20) // 20 MiB

type State struct {
	buff []byte
	idx  int
	err  error
}

func (state *State) fill() error {
	state.idx = 0
	_, state.err = cryptoRand.Read(state.buff)
	if state.err != nil {
		return state.err
	}

	return nil
}

func (state *State) Read(b []byte) (n int, err error) {
	if state.err != nil {
		return 0, state.err
	}

	target := len(b)
	n = 0
	for n != target {
		read := copy(b, state.buff[state.idx:])
		state.idx += read
		n += read
		b = b[read:]

		if state.idx >= len(state.buff) {
			state.err = state.fill()
			if state.err != nil {
				return 0, state.err
			}
		}
	}
	return n, nil
}

func New() (*State, error) {
	state := new(State)
	state.buff = make([]byte, buffSize)
	state.err = state.fill()
	if state.err != nil {
		return nil, state.err
	}

	return state, nil
}
