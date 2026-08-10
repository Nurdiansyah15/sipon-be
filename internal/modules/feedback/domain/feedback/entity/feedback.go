package entity

import (
	"strings"
	"time"

	"sipon-be/internal/modules/feedback/domain/feedback/constant"
	"sipon-be/internal/shared/kernel"
)

type Feedback struct {
	ID             string
	UserID         string
	Title          string
	Body           string
	Category       constant.FeedbackCategory
	IsTakedown     bool
	TakedownReason *string
	TakedownBy     *string
	TakedownAt     *time.Time
	LikeCount      int
	CommentCount   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func NewFeedback(id, userID, title, body string, category constant.FeedbackCategory) (*Feedback, error) {
	f := &Feedback{
		ID:        id,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := f.SetTitle(title); err != nil {
		return nil, err
	}
	if err := f.SetBody(body); err != nil {
		return nil, err
	}
	if err := f.SetCategory(category); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *Feedback) SetTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return kernel.New(constant.CodeFeedbackEmptyTitle)
	}
	f.Title = strings.TrimSpace(title)
	f.UpdatedAt = time.Now()
	return nil
}

func (f *Feedback) SetBody(body string) error {
	if strings.TrimSpace(body) == "" {
		return kernel.New(constant.CodeFeedbackEmptyBody)
	}
	f.Body = body
	f.UpdatedAt = time.Now()
	return nil
}

func (f *Feedback) SetCategory(category constant.FeedbackCategory) error {
	if category == "" {
		f.Category = constant.CategoryLainnya
		f.UpdatedAt = time.Now()
		return nil
	}
	if !constant.ValidCategories[category] {
		return kernel.New(constant.CodeFeedbackInvalidCategory)
	}
	f.Category = category
	f.UpdatedAt = time.Now()
	return nil
}

func (f *Feedback) Update(title, body string, category constant.FeedbackCategory) error {
	if err := f.SetTitle(title); err != nil {
		return err
	}
	if err := f.SetBody(body); err != nil {
		return err
	}
	return f.SetCategory(category)
}

func (f *Feedback) Takedown(adminID string, reason *string) error {
	if f.IsTakedown {
		return kernel.New(constant.CodeFeedbackAlreadyTakedown)
	}
	now := time.Now()
	f.IsTakedown = true
	f.TakedownReason = reason
	f.TakedownBy = &adminID
	f.TakedownAt = &now
	f.UpdatedAt = now
	return nil
}

func (f *Feedback) Restore() error {
	if !f.IsTakedown {
		return kernel.New(constant.CodeFeedbackNotTakedown)
	}
	f.IsTakedown = false
	f.TakedownReason = nil
	f.TakedownBy = nil
	f.TakedownAt = nil
	f.UpdatedAt = time.Now()
	return nil
}

func (f *Feedback) IncrementLike() {
	f.LikeCount++
	f.UpdatedAt = time.Now()
}

func (f *Feedback) DecrementLike() {
	if f.LikeCount > 0 {
		f.LikeCount--
	}
	f.UpdatedAt = time.Now()
}

func (f *Feedback) IncrementComment() {
	f.CommentCount++
	f.UpdatedAt = time.Now()
}

func (f *Feedback) DecrementComment() {
	if f.CommentCount > 0 {
		f.CommentCount--
	}
	f.UpdatedAt = time.Now()
}

func (f *Feedback) SoftDelete() {
	now := time.Now()
	f.DeletedAt = &now
	f.UpdatedAt = now
}
