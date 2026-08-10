package entity

import (
	"time"

	"sipon-be/internal/modules/feedback/domain/attachment/constant"
	"sipon-be/internal/shared/kernel"
)

type Attachment struct {
	ID               string
	FeedbackID       string
	Key              string
	OriginalFilename *string
	MimeType         *string
	Size             *int64
	SortOrder        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewAttachment(id, feedbackID, key string, originalFilename, mimeType *string, size *int64, sortOrder int) (*Attachment, error) {
	if id == "" || feedbackID == "" || key == "" {
		return nil, kernel.New(constant.CodeAttachmentPersistenceFailed)
	}
	now := time.Now()
	return &Attachment{
		ID:               id,
		FeedbackID:       feedbackID,
		Key:              key,
		OriginalFilename: originalFilename,
		MimeType:         mimeType,
		Size:             size,
		SortOrder:        sortOrder,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (a *Attachment) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
}
