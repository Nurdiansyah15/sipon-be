package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/program_transfer_request/constant"
	"sipon-be/internal/shared/kernel"
)

// ProgramTransferRequest adalah permintaan santri untuk pindah program.
// Status: pending → approved / rejected (hanya valid dari pending).
type ProgramTransferRequest struct {
	ID            string
	SantriID      string
	FromProgramID string
	ToProgramID   string
	Status        constant.ProgramTransferRequestStatus
	Notes         *string
	AdminNotes    *string
	ReviewedBy    *string
	ReviewedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// NewProgramTransferRequest membuat request baru berstatus pending.
func NewProgramTransferRequest(id, santriID, fromProgramID, toProgramID string, notes *string) (*ProgramTransferRequest, error) {
	if id == "" || santriID == "" || fromProgramID == "" || toProgramID == "" {
		return nil, kernel.New(constant.CodeProgramTransferRequestNotFound)
	}
	if fromProgramID == toProgramID {
		return nil, kernel.New(constant.CodeProgramTransferRequestSameProgram)
	}
	now := time.Now()
	return &ProgramTransferRequest{
		ID:            id,
		SantriID:      santriID,
		FromProgramID: fromProgramID,
		ToProgramID:   toProgramID,
		Status:        constant.ProgramTransferRequestStatusPending,
		Notes:         notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Approve menyetujui request. Hanya valid dari status pending.
func (r *ProgramTransferRequest) Approve(adminID string) error {
	if r.Status != constant.ProgramTransferRequestStatusPending {
		return kernel.New(constant.CodeProgramTransferRequestInvalidStatus)
	}
	now := time.Now()
	r.Status = constant.ProgramTransferRequestStatusApproved
	r.ReviewedBy = &adminID
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}

// Reject menolak request. Hanya valid dari status pending.
func (r *ProgramTransferRequest) Reject(adminID string, adminNotes *string) error {
	if r.Status != constant.ProgramTransferRequestStatusPending {
		return kernel.New(constant.CodeProgramTransferRequestInvalidStatus)
	}
	now := time.Now()
	r.Status = constant.ProgramTransferRequestStatusRejected
	r.AdminNotes = adminNotes
	r.ReviewedBy = &adminID
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}

func (r *ProgramTransferRequest) SoftDelete() {
	now := time.Now()
	r.DeletedAt = &now
	r.UpdatedAt = now
}
