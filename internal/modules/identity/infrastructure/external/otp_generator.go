package external

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type CryptoOTPGenerator struct{}

func NewCryptoOTPGenerator() *CryptoOTPGenerator {
	return &CryptoOTPGenerator{}
}

func (g *CryptoOTPGenerator) Generate() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
