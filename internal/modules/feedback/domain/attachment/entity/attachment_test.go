package entity_test

import (
	"testing"

	acconstant "sipon-be/internal/modules/feedback/domain/attachment/constant"
	aentity "sipon-be/internal/modules/feedback/domain/attachment/entity"
)

func TestNewAttachment(t *testing.T) {
	of := "bukti.pdf"
	mt := "application/pdf"
	var sz int64 = 1024
	a, err := aentity.NewAttachment("att-1", "fb-1", "feedback/fb-1/uuid.pdf", &of, &mt, &sz, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ID != "att-1" {
		t.Errorf("expected att-1, got %s", a.ID)
	}
	if a.FeedbackID != "fb-1" {
		t.Errorf("expected fb-1, got %s", a.FeedbackID)
	}
	if a.SortOrder != 1 {
		t.Errorf("expected sort_order 1, got %d", a.SortOrder)
	}
}

func TestNewAttachmentValidation(t *testing.T) {
	if _, err := aentity.NewAttachment("", "fb-1", "key", nil, nil, nil, 0); err == nil {
		t.Error("expected error for empty id")
	}
	if _, err := aentity.NewAttachment("att-1", "", "key", nil, nil, nil, 0); err == nil {
		t.Error("expected error for empty feedback id")
	}
	if _, err := aentity.NewAttachment("att-1", "fb-1", "", nil, nil, nil, 0); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestAttachmentSoftDelete(t *testing.T) {
	a, _ := aentity.NewAttachment("att-1", "fb-1", "key", nil, nil, nil, 0)
	a.SoftDelete()
	if a.DeletedAt == nil {
		t.Error("expected DeletedAt set")
	}
}

func TestAllowedContentTypes(t *testing.T) {
	if !acconstant.AllowedContentTypes["image/jpeg"] {
		t.Error("expected image/jpeg allowed")
	}
	if !acconstant.AllowedContentTypes["application/pdf"] {
		t.Error("expected application/pdf allowed")
	}
	if acconstant.AllowedContentTypes["application/x-msdownload"] {
		t.Error("expected executable not allowed")
	}
}
