package googleoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application/ports"
)

type tokenInfoResponse struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified string `json:"email_verified"`
	Error         string `json:"error"`
	ErrorDesc     string `json:"error_description"`
}

type Verifier struct {
	client *http.Client
}

func NewVerifier() *Verifier {
	return &Verifier{client: &http.Client{Timeout: 10 * time.Second}}
}

func (v *Verifier) VerifyIDToken(ctx context.Context, idToken string, allowedClientIDs []string) (*ports.GoogleIdentityInfo, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, fmt.Errorf("id_token wajib diisi")
	}
	if len(allowedClientIDs) == 0 {
		return nil, fmt.Errorf("google client id belum dikonfigurasi")
	}

	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("buat request tokeninfo: %w", err)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request tokeninfo google: %w", err)
	}
	defer resp.Body.Close()

	var payload tokenInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode tokeninfo google: %w", err)
	}

	if resp.StatusCode != http.StatusOK || payload.Error != "" {
		if payload.ErrorDesc != "" {
			return nil, fmt.Errorf("tokeninfo google: %s", payload.ErrorDesc)
		}
		return nil, fmt.Errorf("tokeninfo google tidak valid (http %d)", resp.StatusCode)
	}
	if strings.TrimSpace(payload.Sub) == "" {
		return nil, fmt.Errorf("sub tidak ada pada token")
	}
	if strings.TrimSpace(payload.Email) == "" {
		return nil, fmt.Errorf("email tidak ada pada token")
	}
	if !audienceAllowed(payload.Aud, allowedClientIDs) {
		return nil, fmt.Errorf("audience token tidak diizinkan")
	}

	ev := strings.ToLower(strings.TrimSpace(payload.EmailVerified))
	if ev != "true" && ev != "1" {
		return nil, fmt.Errorf("email google belum terverifikasi")
	}

	return &ports.GoogleIdentityInfo{
		Subject: payload.Sub,
		Email:   strings.ToLower(strings.TrimSpace(payload.Email)),
		Name:    strings.TrimSpace(payload.Name),
		Picture: strings.TrimSpace(payload.Picture),
	}, nil
}

func audienceAllowed(aud string, allowedClientIDs []string) bool {
	for _, id := range allowedClientIDs {
		if aud == id {
			return true
		}
	}
	return false
}
