package external

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"sipon-be/internal/shared/config"
)

func newTestGateway(serverKey string) *MidtransGatewayImpl {
	return NewMidtransGateway(config.MidtransConfig{
		Environment: "sandbox",
		ServerKey:   serverKey,
	})
}

func TestVerifySignatureValid(t *testing.T) {
	g := newTestGateway("SB-Mid-server-testkey")
	orderID := "SIPON-TEST-1"
	statusCode := "200"
	grossAmount := "150000.00"

	hash := sha512.Sum512([]byte("SB-Mid-server-testkey" + orderID + statusCode + grossAmount))
	signatureKey := hex.EncodeToString(hash[:])

	if !g.VerifySignature(orderID, statusCode, grossAmount, signatureKey) {
		t.Error("expected valid signature to pass")
	}
}

func TestVerifySignatureTampered(t *testing.T) {
	g := newTestGateway("SB-Mid-server-testkey")
	orderID := "SIPON-TEST-1"
	statusCode := "200"
	grossAmount := "150000.00"

	hash := sha512.Sum512([]byte("SB-Mid-server-testkey" + orderID + statusCode + grossAmount))
	signatureKey := hex.EncodeToString(hash[:])

	// gross_amount diubah -> signature harus gagal.
	if g.VerifySignature(orderID, statusCode, "150001.00", signatureKey) {
		t.Error("expected tampered payload to fail signature check")
	}
}

func TestVerifySignatureEmptyServerKey(t *testing.T) {
	g := newTestGateway("")
	if g.VerifySignature("order", "200", "1000", "sig") {
		t.Error("expected false when server key is empty")
	}
}

func TestVerifySignatureEmptyFields(t *testing.T) {
	g := newTestGateway("SB-Mid-server-testkey")
	if g.VerifySignature("", "200", "1000", "sig") {
		t.Error("expected false when order id is empty")
	}
}
