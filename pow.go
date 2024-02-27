package pow

import (
	cryptoRand "crypto/rand"
	"errors"
	"hash"
	"io"
	"os"

	"github.com/bernardo1r/pow/internal/rand"

	"golang.org/x/crypto/sha3"
)

type Result struct {
	Challenge []byte
	Digest    []byte
	Zeros     int
}

type Pow struct {
	hash            hash.Hash
	buff            []byte
	inputDigestSize int
	result          *Result
}

func DigestFile(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	hash := sha3.New512()
	n, err := io.Copy(hash, file)
	if err != nil {
		return nil, err
	}
	if n != info.Size() {
		return nil, errors.New("file was not fully read")
	}

	digest := make([]byte, 0, hash.Size())
	return hash.Sum(digest), nil
}

func (p *Pow) redo(state *rand.State) (*Result, error) {
	if state == nil {
		_, err := cryptoRand.Read(p.buff[p.inputDigestSize:])
		if err != nil {
			return nil, err
		}
	} else {
		_, err := state.Read(p.buff[p.inputDigestSize:])
		if err != nil {
			return nil, err
		}
	}

	p.hash.Reset()
	_, err := p.hash.Write(p.buff)
	if err != nil {
		return nil, err
	}

	res := new(Result)
	res.Digest = make([]byte, 0, p.hash.Size())
	res.Digest = p.hash.Sum(res.Digest)
	res.Challenge = make([]byte, len(p.buff[p.inputDigestSize:]))
	copy(res.Challenge, p.buff[p.inputDigestSize:])
	res.Zeros = res.countZeros()
	return res, nil
}

func (res *Result) countZeros() int {
	idx := 0
count:
	for _, b := range res.Digest {
		for i := range 8 {
			if (b>>i)&1 == 1 {
				break count
			}
			idx++
		}
	}
	return idx
}

func New(digest []byte) (*Pow, error) {
	p := new(Pow)
	p.inputDigestSize = len(digest)
	p.hash = sha3.New512()
	p.buff = make([]byte, p.inputDigestSize+p.hash.Size())
	copy(p.buff, digest)
	res, err := p.redo(nil)
	if err != nil {
		return nil, err
	}
	p.result = res

	return p, nil
}

func (p *Pow) Result() *Result {
	return p.result
}

func (p *Pow) Redo(state *rand.State) (*Result, error) {
	res, err := p.redo(state)
	if err != nil {
		return nil, err
	}

	p.result = res
	return res, nil
}
