package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"math/rand"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~")

const (
	CODE_CHALLENGE_METHOD = "S256"
)

func init() {
	rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateCodeVerifer(length int) string {
	if length > 128 {
		length = 128
	}
	if length < 43 {
		length = 43
	}
	return randStringRunes(length)
}

func generateCodeChallenge(code_verifer string) string {
	sum := sha256.Sum256([]byte(code_verifer))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
}

// generate code_verifer (length:43-128) and code_challenge for oauth 2.0 pkce
func NewPKCECode(veriferLength int) (code_verifer, code_challenge string) {
	code_verifer = generateCodeVerifer(veriferLength)
	code_challenge = generateCodeChallenge(code_verifer)
	return
}
