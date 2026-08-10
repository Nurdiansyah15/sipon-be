package entity

import (
	"strings"
	"time"

	"sipon-be/internal/modules/feedback/domain/comment/constant"
	"sipon-be/internal/shared/kernel"
)

type Comment struct {
	ID             string
	FeedbackID     string
	UserID         string
	Body           string
	ReplyToID      *string
	IsTakedown     bool
	TakedownReason *string
	TakedownBy     *string
	TakedownAt     *time.Time
	LikeCount      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func NewComment(id, feedbackID, userID, body string, replyToID *string) (*Comment, error) {
	now := time.Now()
	c := &Comment{
		ID:         id,
		FeedbackID: feedbackID,
		UserID:     userID,
		ReplyToID:  replyToID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := c.SetBody(body); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Comment) SetBody(body string) error {
	if strings.TrimSpace(body) == "" {
		return kernel.New(constant.CodeCommentEmptyBody)
	}
	c.Body = body
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Comment) Update(body string) error {
	return c.SetBody(body)
}

func (c *Comment) Takedown(adminID string, reason *string) error {
	if c.IsTakedown {
		return kernel.New(constant.CodeCommentAlreadyTakedown)
	}
	now := time.Now()
	c.IsTakedown = true
	c.TakedownReason = reason
	c.TakedownBy = &adminID
	c.TakedownAt = &now
	c.UpdatedAt = now
	return nil
}

func (c *Comment) Restore() error {
	if !c.IsTakedown {
		return kernel.New(constant.CodeCommentNotTakedown)
	}
	c.IsTakedown = false
	c.TakedownReason = nil
	c.TakedownBy = nil
	c.TakedownAt = nil
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Comment) IncrementLike() {
	c.LikeCount++
	c.UpdatedAt = time.Now()
}

func (c *Comment) DecrementLike() {
	if c.LikeCount > 0 {
		c.LikeCount--
	}
	c.UpdatedAt = time.Now()
}

func (c *Comment) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
}
