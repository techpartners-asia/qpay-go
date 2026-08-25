package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
)

func GenerateHMAC(secret, data string) (sha string) {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandStringBytesMaskImprSrcSB returns a random string of n letters.
//
// Values produced here end up in payment identifiers and callback
// parameters, so they are drawn from crypto/rand: a time-seeded math/rand
// stream is both predictable by an attacker who can observe one value and
// unsafe to share between goroutines.
func RandStringBytesMaskImprSrcSB(n int) string {
	if n <= 0 {
		return ""
	}

	sb := strings.Builder{}
	sb.Grow(n)

	max := big.NewInt(int64(len(letterBytes)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			// crypto/rand.Reader never fails on supported platforms; if the
			// system entropy source is broken, returning a guessable token
			// would be worse than returning nothing.
			return ""
		}
		sb.WriteByte(letterBytes[idx.Int64()])
	}

	return sb.String()
}
