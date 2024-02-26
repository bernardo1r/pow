package pow

import (
	"crypto/rand"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/sha3"
)

type Result struct {
	Challenge []byte
	Digest    []byte
	Zeros     int
}

type Pow struct {
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

func (p *Pow) redo() (*Result, error) {
	_, err := rand.Read(p.buff[p.inputDigestSize:])
	if err != nil {
		return nil, err
	}

	res := new(Result)
	final := sha3.Sum512(p.buff)
	res.Digest = final[:]
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
	p.buff = make([]byte, p.inputDigestSize+sha3.New512().Size())
	copy(p.buff, digest)
	res, err := p.redo()
	if err != nil {
		return nil, err
	}
	p.result = res

	return p, nil
}

func (p *Pow) Result() *Result {
	return p.result
}

func (p *Pow) Redo() (*Result, error) {
	res, err := p.redo()
	if err != nil {
		return nil, err
	}

	p.result = res
	return res, nil
}
