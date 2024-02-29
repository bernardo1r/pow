package pow

import (
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
	state           *rand.State
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

func (p *Pow) redo(res *Result) (*Result, error) {
	_, err := p.state.Read(p.buff[p.inputDigestSize:])
	if err != nil {
		return nil, err
	}

	p.hash.Reset()
	_, err = p.hash.Write(p.buff)
	if err != nil {
		return nil, err
	}

	if res == nil {
		res = new(Result)
		res.Challenge = make([]byte, p.hash.Size())
		res.Digest = make([]byte, 0, p.hash.Size())
	} else {
		res.Digest = res.Digest[:0]
	}

	res.Digest = p.hash.Sum(res.Digest)
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
	var err error
	p.state, err = rand.New()
	if err != nil {
		return nil, err
	}
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

func (p *Pow) Redo(res *Result) (*Result, error) {
	res, err := p.redo(res)
	if err != nil {
		return nil, err
	}

	p.result = res
	return res, nil
}
