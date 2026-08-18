package mq

import "errors"

// FingerprintSyncPayload adalah DTO payload untuk akademik.fingerprint.sync.
type FingerprintSyncPayload struct {
	SessionID string `json:"session_id"`
}

func (p FingerprintSyncPayload) Validate() error {
	if p.SessionID == "" {
		return errors.New("session_id wajib diisi")
	}
	return nil
}

// SessionAutoClosePayload adalah DTO payload untuk akademik.session.auto_close.
type SessionAutoClosePayload struct {
	SessionID string `json:"session_id"`
}

func (p SessionAutoClosePayload) Validate() error {
	if p.SessionID == "" {
		return errors.New("session_id wajib diisi")
	}
	return nil
}
