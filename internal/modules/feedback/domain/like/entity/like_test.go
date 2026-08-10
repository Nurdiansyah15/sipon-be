package entity_test

import (
	"testing"

	lconstant "sipon-be/internal/modules/feedback/domain/like/constant"
	lentity "sipon-be/internal/modules/feedback/domain/like/entity"
)

func TestNewLike(t *testing.T) {
	l := lentity.NewLike("like-1", "user-1", lconstant.TargetFeedback, "fb-1")
	if l.ID != "like-1" {
		t.Errorf("expected like-1, got %s", l.ID)
	}
	if l.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", l.UserID)
	}
	if l.TargetType != lconstant.TargetFeedback {
		t.Errorf("expected target feedback, got %s", l.TargetType)
	}
	if l.TargetID != "fb-1" {
		t.Errorf("expected fb-1, got %s", l.TargetID)
	}
	if l.CreatedAt.IsZero() {
		t.Error("expected CreatedAt set")
	}
}

func TestNewLikeCommentTarget(t *testing.T) {
	l := lentity.NewLike("like-2", "user-1", lconstant.TargetComment, "comment-1")
	if l.TargetType != lconstant.TargetComment {
		t.Errorf("expected target comment, got %s", l.TargetType)
	}
}
