package session

import (
	"crypto/rand"
	"math/big"
	"time"
)

const ulidAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func newULID(now time.Time) (string, error) {
	var value [16]byte
	ms := uint64(now.UnixMilli())
	value[0] = byte(ms >> 40)
	value[1] = byte(ms >> 32)
	value[2] = byte(ms >> 24)
	value[3] = byte(ms >> 16)
	value[4] = byte(ms >> 8)
	value[5] = byte(ms)
	if _, err := rand.Read(value[6:]); err != nil {
		return "", err
	}

	// ULID encodes 128 bits in 26 Crockford base32 characters. Repeated
	// division keeps the leading two padding bits implicit in the first digit.
	var out [26]byte
	n := new(big.Int).SetBytes(value[:])
	base := big.NewInt(32)
	rem := new(big.Int)
	for i := len(out) - 1; i >= 0; i-- {
		n.QuoRem(n, base, rem)
		out[i] = ulidAlphabet[rem.Int64()]
	}
	return string(out[:]), nil
}

func isULID(value string) bool {
	if len(value) != 26 || value[0] > '7' {
		return false
	}
	for _, c := range value {
		valid := (c >= '0' && c <= '9') || (c >= 'A' && c <= 'H') ||
			(c >= 'J' && c <= 'K') || (c >= 'M' && c <= 'N') ||
			(c >= 'P' && c <= 'T') || (c >= 'V' && c <= 'Z')
		if !valid {
			return false
		}
	}
	return true
}
