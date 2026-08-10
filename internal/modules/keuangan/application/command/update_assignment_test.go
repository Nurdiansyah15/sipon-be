package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsEntity "sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	"sipon-be/internal/shared/kernel"
)

type fakeAssignmentRepo struct {
	byID     map[string]*bsEntity.SantriBillingAssignment
	overlap  bool
	findErr  error
	updated  *bsEntity.SantriBillingAssignment
}

func (f *fakeAssignmentRepo) Save(ctx context.Context, a *bsEntity.SantriBillingAssignment) error {
	return nil
}
func (f *fakeAssignmentRepo) Update(ctx context.Context, a *bsEntity.SantriBillingAssignment) error {
	f.updated = a
	return nil
}
func (f *fakeAssignmentRepo) FindByID(ctx context.Context, id string) (*bsEntity.SantriBillingAssignment, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if a, ok := f.byID[id]; ok {
		return a, nil
	}
	return nil, kernel.WrapMsg(bsConst.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
}
func (f *fakeAssignmentRepo) FindActiveBySantriID(ctx context.Context, santriID string) (*bsEntity.SantriBillingAssignment, error) {
	return nil, kernel.WrapMsg(bsConst.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
}
func (f *fakeAssignmentRepo) FindActiveBySantriIDAt(ctx context.Context, santriID string, atDate time.Time) (*bsEntity.SantriBillingAssignment, error) {
	return nil, kernel.WrapMsg(bsConst.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
}
func (f *fakeAssignmentRepo) EndAssignment(ctx context.Context, id string, effectiveUntil time.Time) error {
	return nil
}
func (f *fakeAssignmentRepo) ListBySantriID(ctx context.Context, santriID string) ([]*bsEntity.SantriBillingAssignment, error) {
	return nil, nil
}
func (f *fakeAssignmentRepo) HasOverlappingAssignment(ctx context.Context, santriID string, from time.Time, until *time.Time, excludeID string) (bool, error) {
	return f.overlap, nil
}

type fakeSchemeRepo struct {
	scheme *bsEntity.BillingScheme
}

func (f *fakeSchemeRepo) Save(ctx context.Context, s *bsEntity.BillingScheme) error { return nil }
func (f *fakeSchemeRepo) Update(ctx context.Context, s *bsEntity.BillingScheme) error {
	return nil
}
func (f *fakeSchemeRepo) FindByID(ctx context.Context, id string) (*bsEntity.BillingScheme, error) {
	if f.scheme != nil && f.scheme.ID == id {
		return f.scheme, nil
	}
	return nil, kernel.WrapMsg(bsConst.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
}
func (f *fakeSchemeRepo) List(ctx context.Context, q bsRepo.BillingSchemeListQuery) (*bsRepo.BillingSchemeListResult, error) {
	return &bsRepo.BillingSchemeListResult{}, nil
}
func (f *fakeSchemeRepo) AddItems(ctx context.Context, schemeID string, items []*bsEntity.BillingSchemeItem) error {
	return nil
}
func (f *fakeSchemeRepo) RemoveItem(ctx context.Context, schemeID, itemID string) error { return nil }

func TestUpdateAssignment_Success(t *testing.T) {
	assignment := &bsEntity.SantriBillingAssignment{
		ID:              "assign-1",
		SantriID:        "santri-1",
		BillingSchemeID: "scheme-old",
		EffectiveFrom:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	scheme, _ := bsEntity.NewBillingScheme("scheme-new", "Skema Baru", "user-1")

	assignmentRepo := &fakeAssignmentRepo{byID: map[string]*bsEntity.SantriBillingAssignment{"assign-1": assignment}}
	uc := NewUpdateAssignmentUseCase(assignmentRepo, &fakeSchemeRepo{scheme: scheme})

	_, err := uc.Execute(context.Background(), "assign-1", dto.UpdateAssignmentRequest{
		BillingSchemeID: "scheme-new",
		EffectiveFrom:   "2026-02-01",
		EffectiveUntil:  ptr("2026-06-30"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assignmentRepo.updated == nil {
		t.Fatal("expected assignment to be updated")
	}
	if assignmentRepo.updated.BillingSchemeID != "scheme-new" {
		t.Errorf("expected scheme-new, got %s", assignmentRepo.updated.BillingSchemeID)
	}
	if assignmentRepo.updated.EffectiveFrom.Format("2006-01-02") != "2026-02-01" {
		t.Errorf("expected effective_from 2026-02-01, got %v", assignmentRepo.updated.EffectiveFrom)
	}
	if assignmentRepo.updated.EffectiveUntil == nil || assignmentRepo.updated.EffectiveUntil.Format("2006-01-02") != "2026-06-30" {
		t.Errorf("expected effective_until 2026-06-30, got %v", assignmentRepo.updated.EffectiveUntil)
	}
}

func TestUpdateAssignment_NotFound(t *testing.T) {
	assignmentRepo := &fakeAssignmentRepo{byID: map[string]*bsEntity.SantriBillingAssignment{}}
	uc := NewUpdateAssignmentUseCase(assignmentRepo, &fakeSchemeRepo{})

	_, err := uc.Execute(context.Background(), "missing", dto.UpdateAssignmentRequest{
		BillingSchemeID: "scheme-1",
		EffectiveFrom:   "2026-02-01",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kernel.AppError, got %T: %v", err, err)
	}
	if ke.Code != application.ErrCodeNotFound {
		t.Fatalf("expected code %s, got %s (%v)", application.ErrCodeNotFound, ke.Code, err)
	}
}

func TestUpdateAssignment_SchemeNotFound(t *testing.T) {
	assignment := &bsEntity.SantriBillingAssignment{ID: "assign-1", SantriID: "santri-1"}
	assignmentRepo := &fakeAssignmentRepo{byID: map[string]*bsEntity.SantriBillingAssignment{"assign-1": assignment}}
	uc := NewUpdateAssignmentUseCase(assignmentRepo, &fakeSchemeRepo{})

	_, err := uc.Execute(context.Background(), "assign-1", dto.UpdateAssignmentRequest{
		BillingSchemeID: "scheme-missing",
		EffectiveFrom:   "2026-02-01",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kernel.AppError, got %T: %v", err, err)
	}
	if ke.Code != application.ErrCodeNotFound {
		t.Fatalf("expected code %s, got %s (%v)", application.ErrCodeNotFound, ke.Code, err)
	}
}

func TestUpdateAssignment_Overlap(t *testing.T) {
	assignment := &bsEntity.SantriBillingAssignment{ID: "assign-1", SantriID: "santri-1"}
	scheme, _ := bsEntity.NewBillingScheme("scheme-1", "Skema", "user-1")
	assignmentRepo := &fakeAssignmentRepo{
		byID:    map[string]*bsEntity.SantriBillingAssignment{"assign-1": assignment},
		overlap: true,
	}
	uc := NewUpdateAssignmentUseCase(assignmentRepo, &fakeSchemeRepo{scheme: scheme})

	_, err := uc.Execute(context.Background(), "assign-1", dto.UpdateAssignmentRequest{
		BillingSchemeID: "scheme-1",
		EffectiveFrom:   "2026-02-01",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kernel.AppError, got %T: %v", err, err)
	}
	if ke.Code != application.ErrCodeConflict {
		t.Fatalf("expected code %s, got %s (%v)", application.ErrCodeConflict, ke.Code, err)
	}
}

func ptr(s string) *string {
	return &s
}
