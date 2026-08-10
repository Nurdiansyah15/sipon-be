package dto

type KeuanganSettingResponse struct {
	DefaultPaymentDebitAccountID *string               `json:"default_payment_debit_account_id,omitempty"`
	DefaultPaymentDebitAccount   *AccountBriefResponse `json:"default_payment_debit_account,omitempty"`
}

type UpdateKeuanganSettingRequest struct {
	DefaultPaymentDebitAccountID *string `json:"default_payment_debit_account_id"`
}
