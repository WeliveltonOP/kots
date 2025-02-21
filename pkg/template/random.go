package template

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"regexp/syntax"

	"github.com/pkg/errors"
)

const (
	DefaultCharset = "[_A-Za-z0-9]"
)

// stolen from https://github.com/replicatedhq/replicated/blob/8ce3ed40436e38b8089387d103623dbe09bbf1c0/pkg/commands/random.go#L22
func (ctx *StaticCtx) RandomString(length uint64, providedCharset ...string) string {
	charset := DefaultCharset
	if len(providedCharset) >= 1 {
		charset = providedCharset[0]
	}
	regExp, err := syntax.Parse(charset, syntax.Perl)
	if err != nil {
		return ""
	}

	regExp = regExp.Simplify()
	var b bytes.Buffer
	for i := 0; i < int(length); i++ {
		if err := ctx.genString(&b, regExp); err != nil {
			return ""
		}
	}

	result := b.String()
	return result
}

func (ctx *StaticCtx) genString(w *bytes.Buffer, rx *syntax.Regexp) error {
	switch rx.Op {
	case syntax.OpCharClass:
		sum := 0
		for i := 0; i < len(rx.Rune); i += 2 {
			sum += 1 + int(rx.Rune[i+1]-rx.Rune[i])
		}

		for i, nth := 0, rune(randint(sum)); i < len(rx.Rune); i += 2 {
			min, max := rx.Rune[i], rx.Rune[i+1]
			delta := max - min
			if nth <= delta {
				w.WriteRune(min + nth)
				return nil
			}
			nth -= 1 + delta
		}
	default:
		return errors.New("invalid charset")
	}

	return nil
}

func randint(max int) int {
	var bigmax big.Int
	bigmax.SetInt64(int64(max))

	res, err := rand.Int(rand.Reader, &bigmax)
	if err != nil {
		panic(err)
	}

	return int(res.Int64())
}

// RandomBytes returns a base64-encoded byte array allowing the full range of byte values.
func (ctx *StaticCtx) RandomBytes(length uint64) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf)
}
