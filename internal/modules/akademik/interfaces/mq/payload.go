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

// SessionAutoOpenPayload adalah DTO payload untuk akademik.session.auto_open.
type SessionAutoOpenPayload struct {
	SessionID string `json:"session_id"`
}

func (p SessionAutoOpenPayload) Validate() error {
	if p.SessionID == "" {
		return errors.New("session_id wajib diisi")
	}
	return nil
}

// SessionReminderPayload adalah DTO payload untuk akademik.session.reminder.
type SessionReminderPayload struct {
	SessionID string `json:"session_id"`
}

func (p SessionReminderPayload) Validate() error {
	if p.SessionID == "" {
		return errors.New("session_id wajib diisi")
	}
	return nil
}
