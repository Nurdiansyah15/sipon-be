package entity_test

import (
	"testing"

	"sipon-be/internal/modules/feedback/domain/feedback/constant"
	"sipon-be/internal/modules/feedback/domain/feedback/entity"
	"sipon-be/internal/shared/kernel"
)

func createTestFeedback() *entity.Feedback {
	f, _ := entity.NewFeedback("fb-1", "user-1", "Judul", "Isi feedback", constant.CategorySaran)
	return f
}

func TestNewFeedback(t *testing.T) {
	f := createTestFeedback()
	if f.ID != "fb-1" {
		t.Errorf("expected id fb-1, got %s", f.ID)
	}
	if f.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", f.UserID)
	}
	if f.Category != constant.CategorySaran {
		t.Errorf("expected category saran, got %s", f.Category)
	}
	if f.LikeCount != 0 || f.CommentCount != 0 {
		t.Error("new feedback should start with zero counters")
	}
}

func TestNewFeedbackValidation(t *testing.T) {
	if _, err := entity.NewFeedback("id", "user", "", "body", constant.CategoryLainnya); err == nil {
		t.Error("expected error for empty title")
	}
	if _, err := entity.NewFeedback("id", "user", "title", "", constant.CategoryLainnya); err == nil {
		t.Error("expected error for empty body")
	}
	if _, err := entity.NewFeedback("id", "user", "title", "body", constant.FeedbackCategory("invalid")); err == nil {
		t.Error("expected error for invalid category")
	}
}

func TestNewFeedbackDefaultCategory(t *testing.T) {
	f, err := entity.NewFeedback("id", "user", "title", "body", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Category != constant.CategoryLainnya {
		t.Errorf("expected default category lainnya, got %s", f.Category)
	}
}

func TestFeedbackUpdate(t *testing.T) {
	f := createTestFeedback()
	if err := f.Update("Judul baru", "Isi baru", constant.CategoryPengaduan); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Title != "Judul baru" || f.Body != "Isi baru" {
		t.Error("expected updated title and body")
	}
	if f.Category != constant.CategoryPengaduan {
		t.Errorf("expected category pengaduan, got %s", f.Category)
	}
}

func TestFeedbackTakedownRestore(t *testing.T) {
	f := createTestFeedback()
	reason := "tidak pantas"
	if err := f.Takedown("admin-1", &reason); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.IsTakedown {
		t.Error("expected feedback to be takedown")
	}
	if f.TakedownBy == nil || *f.TakedownBy != "admin-1" {
		t.Error("expected takedown_by admin-1")
	}
	if f.TakedownAt == nil {
		t.Error("expected takedown_at set")
	}

	if err := f.Takedown("admin-1", &reason); err == nil {
		t.Error("expected error when takedown twice")
	}

	if err := f.Restore(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.IsTakedown {
		t.Error("expected feedback restored")
	}
	if f.TakedownReason != nil || f.TakedownBy != nil || f.TakedownAt != nil {
		t.Error("expected takedown fields cleared after restore")
	}

	if err := f.Restore(); err == nil {
		t.Error("expected error when restoring non-takedown feedback")
	}
}

func TestFeedbackLikeCounters(t *testing.T) {
	f := createTestFeedback()
	f.IncrementLike()
	f.IncrementLike()
	if f.LikeCount != 2 {
		t.Errorf("expected like_count 2, got %d", f.LikeCount)
	}
	f.DecrementLike()
	f.DecrementLike()
	f.DecrementLike() // should not go negative
	if f.LikeCount != 0 {
		t.Errorf("expected like_count 0, got %d", f.LikeCount)
	}
}

func TestFeedbackCommentCounters(t *testing.T) {
	f := createTestFeedback()
	f.IncrementComment()
	if f.CommentCount != 1 {
		t.Errorf("expected comment_count 1, got %d", f.CommentCount)
	}
	f.DecrementComment()
	f.DecrementComment() // should not go negative
	if f.CommentCount != 0 {
		t.Errorf("expected comment_count 0, got %d", f.CommentCount)
	}
}

func TestFeedbackSoftDelete(t *testing.T) {
	f := createTestFeedback()
	f.SoftDelete()
	if f.DeletedAt == nil {
		t.Error("expected DeletedAt set")
	}
}

func TestFeedbackEmptyTitleDomainCode(t *testing.T) {
	_, err := entity.NewFeedback("id", "user", "  ", "body", constant.CategoryLainnya)
	if err == nil {
		t.Fatal("expected error")
	}
	var ke *kernel.AppError
	if !asAppError(err, &ke) {
		t.Fatalf("expected kernel.AppError, got %T", err)
	}
	if ke.Code != constant.CodeFeedbackEmptyTitle {
		t.Errorf("expected %s, got %s", constant.CodeFeedbackEmptyTitle, ke.Code)
	}
}
