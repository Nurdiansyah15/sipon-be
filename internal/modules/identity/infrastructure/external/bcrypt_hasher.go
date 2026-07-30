package external

import (
	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *BcryptPasswordHasher) Verify(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
