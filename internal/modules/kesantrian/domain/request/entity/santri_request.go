package entity

import (
	"time"

	"sipon-be/internal/modules/kesantrian/domain/request/constant"
	"sipon-be/internal/shared/kernel"
)

type SantriRequest struct {
	ID     string
	UserID string
	NIS    *string
	Status constant.SantriRequestStatus
	Notes  *string

	ReviewedBy *string
	ReviewedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewSantriRequest(id, userID string) (*SantriRequest, error) {
	if id == "" || userID == "" {
		return nil, kernel.New(constant.CodeSantriRequestNotFound)
	}
	now := time.Now()
	return &SantriRequest{
		ID:        id,
		UserID:    userID,
		Status:    constant.SantriRequestStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Approve and Reject are only valid from status=pending, mirroring
// sipon-api's guard.
func (r *SantriRequest) Approve(reviewerID, nis string) error {
	if r.Status != constant.SantriRequestStatusPending {
		return kernel.New(constant.CodeSantriRequestInvalidStatus)
	}
	now := time.Now()
	r.Status = constant.SantriRequestStatusApproved
	r.NIS = &nis
	r.ReviewedBy = &reviewerID
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}

func (r *SantriRequest) Reject(reviewerID string, notes *string) error {
	if r.Status != constant.SantriRequestStatusPending {
		return kernel.New(constant.CodeSantriRequestInvalidStatus)
	}
	now := time.Now()
	r.Status = constant.SantriRequestStatusRejected
	r.Notes = notes
	r.ReviewedBy = &reviewerID
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}
