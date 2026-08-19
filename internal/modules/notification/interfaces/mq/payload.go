package mq

import "errors"

type LoginSucceededPayload struct {
	UserID string `json:"user_id"`
}

func (p LoginSucceededPayload) Validate() error {
	if p.UserID == "" {
		return errors.New("user_id wajib diisi")
	}
	return nil
}
