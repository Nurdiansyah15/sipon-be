package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type FonnteSMSSender struct {
	client *http.Client
	token  string
	apiURL string
}

func NewFonnteSMSSender(token, apiURL string) *FonnteSMSSender {
	return &FonnteSMSSender{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		token:  token,
		apiURL: apiURL,
	}
}

type fonnteRequest struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}

type fonnteResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func (s *FonnteSMSSender) SendOTP(toPhone, otp string) error {
	if s.token == "" {
		return nil
	}

	message := fmt.Sprintf("Your Sipon OTP code is: %s. Valid for 10 minutes.", otp)

	reqBody, err := json.Marshal(fonnteRequest{
		Target:  toPhone,
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("marshal fonnte request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create fonnte request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("fonnte request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read fonnte response: %w", err)
	}

	var result fonnteResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("unmarshal fonnte response: %w", err)
	}

	if !result.Status {
		return fmt.Errorf("fonnte error: %s", result.Message)
	}

	return nil
}
