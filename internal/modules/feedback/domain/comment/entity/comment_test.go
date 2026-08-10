package entity_test

import (
	"errors"
	"testing"

	"sipon-be/internal/modules/feedback/domain/comment/constant"
	"sipon-be/internal/modules/feedback/domain/comment/entity"
	"sipon-be/internal/shared/kernel"
)

func asAppError(err error, target **kernel.AppError) bool {
	return errors.As(err, target)
}

func createTestComment() *entity.Comment {
	replyTo := "comment-0"
	c, _ := entity.NewComment("comment-1", "fb-1", "user-1", "Komentar", &replyTo)
	return c
}

func TestNewComment(t *testing.T) {
	c := createTestComment()
	if c.ID != "comment-1" {
		t.Errorf("expected comment-1, got %s", c.ID)
	}
	if c.FeedbackID != "fb-1" {
		t.Errorf("expected fb-1, got %s", c.FeedbackID)
	}
	if c.ReplyToID == nil || *c.ReplyToID != "comment-0" {
		t.Error("expected reply_to_id comment-0")
	}
}

func TestNewCommentValidation(t *testing.T) {
	if _, err := entity.NewComment("id", "fb-1", "user-1", "   ", nil); err == nil {
		t.Error("expected error for empty body")
	}
}

func TestCommentUpdate(t *testing.T) {
	c := createTestComment()
	if err := c.Update("Komentar baru"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Body != "Komentar baru" {
		t.Errorf("expected updated body, got %s", c.Body)
	}
}

func TestCommentTakedownRestore(t *testing.T) {
	c := createTestComment()
	reason := "tidak pantas"
	if err := c.Takedown("admin-1", &reason); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.IsTakedown {
		t.Error("expected comment takedown")
	}
	if err := c.Takedown("admin-1", &reason); err == nil {
		t.Error("expected error when takedown twice")
	}
	if err := c.Restore(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.IsTakedown {
		t.Error("expected comment restored")
	}
	if err := c.Restore(); err == nil {
		t.Error("expected error when restoring non-takedown comment")
	}
}

func TestCommentLikeCounters(t *testing.T) {
	c := createTestComment()
	c.IncrementLike()
	if c.LikeCount != 1 {
		t.Errorf("expected like_count 1, got %d", c.LikeCount)
	}
	c.DecrementLike()
	c.DecrementLike() // should not go negative
	if c.LikeCount != 0 {
		t.Errorf("expected like_count 0, got %d", c.LikeCount)
	}
}

func TestCommentSoftDelete(t *testing.T) {
	c := createTestComment()
	c.SoftDelete()
	if c.DeletedAt == nil {
		t.Error("expected DeletedAt set")
	}
}

func TestNewCommentEmptyBodyCode(t *testing.T) {
	_, err := entity.NewComment("id", "fb-1", "user-1", "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var ke *kernel.AppError
	if !asAppError(err, &ke) {
		t.Fatalf("expected kernel.AppError, got %T", err)
	}
	if ke.Code != constant.CodeCommentEmptyBody {
		t.Errorf("expected %s, got %s", constant.CodeCommentEmptyBody, ke.Code)
	}
}
