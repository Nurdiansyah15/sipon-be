package external

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sipon-be/internal/modules/keuangan/application/ports"
	pgConst "sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/kernel"
)

const (
	midtransSnapSandboxBase = "https://app.sandbox.midtrans.com/snap/v1"
	midtransSnapProdBase    = "https://app.midtrans.com/snap/v1"
	midtransAPISandboxBase  = "https://api.sandbox.midtrans.com/v2"
	midtransAPIProdBase     = "https://api.midtrans.com/v2"
)

// MidtransGatewayImpl mengimplementasikan ports.MidtransGateway memakai
// standard library HTTP client dengan Basic Auth dari server key.
type MidtransGatewayImpl struct {
	client    *http.Client
	serverKey string
	snapBase  string
	apiBase   string
}

func NewMidtransGateway(cfg config.MidtransConfig) *MidtransGatewayImpl {
	snapBase := cfg.SnapBaseURL
	if snapBase == "" {
		if cfg.Environment == "production" {
			snapBase = midtransSnapProdBase
		} else {
			snapBase = midtransSnapSandboxBase
		}
	}
	apiBase := cfg.APIBaseURL
	if apiBase == "" {
		if cfg.Environment == "production" {
			apiBase = midtransAPIProdBase
		} else {
			apiBase = midtransAPISandboxBase
		}
	}
	return &MidtransGatewayImpl{
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
		serverKey: cfg.ServerKey,
		snapBase:  snapBase,
		apiBase:   apiBase,
	}
}

type snapTransactionPayload struct {
	TransactionDetails struct {
		OrderID     string  `json:"order_id"`
		GrossAmount float64 `json:"gross_amount"`
	} `json:"transaction_details"`
	ItemDetails     []snapItemPayload `json:"item_details,omitempty"`
	CustomerDetails struct {
		FirstName string `json:"first_name,omitempty"`
		Email     string `json:"email,omitempty"`
		Phone     string `json:"phone,omitempty"`
	} `json:"customer_details,omitempty"`
	Expiry struct {
		Unit     string `json:"unit"`
		Duration int    `json:"duration"`
	} `json:"expiry,omitempty"`
}

type snapItemPayload struct {
	ID       string  `json:"id"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Name     string  `json:"name"`
}

type snapTransactionResponse struct {
	Token         string `json:"token"`
	RedirectURL   string `json:"redirect_url"`
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_message"`
}

func (g *MidtransGatewayImpl) CreateSnapTransaction(ctx context.Context, req ports.SnapTransactionRequest) (*ports.SnapTransactionResponse, error) {
	if g.serverKey == "" {
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, "server key Midtrans belum dikonfigurasi", nil)
	}

	payload := snapTransactionPayload{}
	payload.TransactionDetails.OrderID = req.OrderID
	payload.TransactionDetails.GrossAmount = req.GrossAmount
	if len(req.Items) > 0 {
		payload.ItemDetails = make([]snapItemPayload, 0, len(req.Items))
		for _, it := range req.Items {
			payload.ItemDetails = append(payload.ItemDetails, snapItemPayload{
				ID:       it.ID,
				Price:    it.Price,
				Quantity: it.Quantity,
				Name:     it.Name,
			})
		}
	}
	if req.CustomerName != "" || req.CustomerEmail != "" || req.CustomerPhone != "" {
		payload.CustomerDetails.FirstName = req.CustomerName
		payload.CustomerDetails.Email = req.CustomerEmail
		payload.CustomerDetails.Phone = req.CustomerPhone
	}
	if req.ExpiryMinutes > 0 {
		payload.Expiry.Unit = "minutes"
		payload.Expiry.Duration = req.ExpiryMinutes
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, "gagal menyusun request Midtrans", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.snapBase+"/transactions", bytes.NewReader(body))
	if err != nil {
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, "gagal membuat request Midtrans", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+g.serverKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, "gagal menghubungi Midtrans", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, "gagal membaca response Midtrans", err)
	}

	var result snapTransactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, "gagal membaca response Midtrans", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		msg := result.StatusMessage
		if msg == "" {
			msg = string(respBody)
		}
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, fmt.Sprintf("Midtrans menolak transaksi: %s", msg), nil)
	}
	if result.Token == "" {
		return nil, kernel.WrapMsg(pgConst.CodePaymentGatewayAPIFailed, "Midtrans tidak mengembalikan snap token", nil)
	}

	return &ports.SnapTransactionResponse{
		Token:       result.Token,
		RedirectURL: result.RedirectURL,
	}, nil
}

// VerifySignature memverifikasi signature_key notifikasi Midtrans sesuai
// standar resmi: SHA512(server_key + order_id + status_code + gross_amount).
// Selalu aman dipanggil (return false) bila server key kosong.
func (g *MidtransGatewayImpl) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	if g.serverKey == "" || orderID == "" || statusCode == "" || grossAmount == "" || signatureKey == "" {
		return false
	}
	hash := sha512.Sum512([]byte(g.serverKey + orderID + statusCode + grossAmount))
	expected := hex.EncodeToString(hash[:])
	return expected == signatureKey
}
