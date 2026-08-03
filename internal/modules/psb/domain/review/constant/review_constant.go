package constant

import "sipon-be/internal/shared/kernel"

type ReviewStage string

const (
	StagePendaftaran ReviewStage = "pendaftaran"
	StageDaftarUlang ReviewStage = "daftar_ulang"
)

type ReviewAction string

const (
	ActionPerluRevisi ReviewAction = "perlu_revisi"
	ActionDitolak     ReviewAction = "ditolak"
	ActionDiterima    ReviewAction = "diterima"
)

const (
	ErrCodeInvalidReview kernel.Code = "INVALID_REVIEW"
)
