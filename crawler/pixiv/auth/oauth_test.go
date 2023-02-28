package auth

import (
	"log"
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	log.Println(NewPKCECode(32))
}
