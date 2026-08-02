package dto

// ImportSantriResultItem reports the outcome of a single row from an
// uploaded santri-import spreadsheet.
type ImportSantriResultItem struct {
	RowNumber         int    `json:"row_number"`
	NIS               string `json:"nis"`
	Status            string `json:"status"` // "success" | "error"
	Message           string `json:"message,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	SantriID          string `json:"santri_id,omitempty"`
	GeneratedPassword string `json:"generated_password,omitempty"`
}

type ImportSantriResponse struct {
	Items        []ImportSantriResultItem `json:"items"`
	SuccessCount int                      `json:"success_count"`
	ErrorCount   int                      `json:"error_count"`
}
