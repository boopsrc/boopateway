package security

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/boopateway/boopateway/internal/domain/gateway"
)

type BcryptCodeHasher struct{}

var _ gateway.CodeHasher = BcryptCodeHasher{}

func (BcryptCodeHasher) Hash(code string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
}

func (BcryptCodeHasher) Compare(hash []byte, code string) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(code)) == nil
}
