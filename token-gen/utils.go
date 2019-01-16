package tg

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultAlphabet    = "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NumberAlphabet     = "1234567890"
	DefaultTokenLength = 64
)

var (
	Default = NewGenerator(DefaultAlphabet)
)

type tokenGenerator struct {
	alphabet      []byte
	src           sync.Pool
	letterIdxBits uint
	letterIdxMask int64
	letterIdxMax  uint
}

func NewGenerator(alphabet string) *tokenGenerator {
	bytes := []byte(alphabet)
	l := len(bytes)
	letterIdxBits := uint(len(strconv.FormatInt(int64(l), 2)))
	return &tokenGenerator{
		alphabet: bytes,
		src: sync.Pool{New: func() interface{} {
			return rand.NewSource(time.Now().UnixNano())
		}},
		letterIdxBits: letterIdxBits,
		letterIdxMask: 1<<letterIdxBits - 1,
		letterIdxMax:  63 / letterIdxBits,
	}
}

func (gen *tokenGenerator) NextDefault() string {
	return gen.Next(DefaultTokenLength)
}

func (gen *tokenGenerator) Next(tokenLength int) string {
	b := make([]byte, tokenLength)
	s := gen.src.Get().(rand.Source)

	for i, cache, remain := tokenLength-1, s.Int63(), gen.letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = s.Int63(), gen.letterIdxMax
		}
		if idx := int(cache & gen.letterIdxMask); idx < len(gen.alphabet) {
			b[i] = gen.alphabet[idx]
			i--
		}
		cache >>= gen.letterIdxBits
		remain--
	}

	gen.src.Put(s)

	return string(b)
}
