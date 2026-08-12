package command

import (
	"context"
	"testing"

	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progEntity "sipon-be/internal/modules/akademik/domain/program/entity"
	spEntity "sipon-be/internal/modules/akademik/domain/santri_program/entity"
)

func strP(s string) *string { return &s }

// --- Admin assign program ---

func TestAssignSantriProgramAdminNewProgram(t *testing.T) {
	prog := &progEntity.Program{ID: "prog-2", Code: "KITAB", Name: "Kitab", Status: "active"}
	programRepo := &fakeProgramRepo{}
	programRepo.programByID = prog

	santriProgramRepo := &fakeSantriProgramRepo{
		activeBySantri: &spEntity.SantriProgram{ID: "sp-1", SantriID: "santri-1", ProgramID: "prog-1", IsActive: true},
	}

	uc := NewAssignSantriProgramAdminUseCase(santriProgramRepo, programRepo, fakeTransactor{})
	resp, err := uc.Execute(context.Background(), "santri-1", "prog-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ProgramID != "prog-2" {
		t.Fatalf("expected program prog-2, got %s", resp.ProgramID)
	}
	if resp.Program.Code != "KITAB" {
		t.Fatalf("expected KITAB, got %s", resp.Program.Code)
	}
	if !santriProgramRepo.deactivated {
		t.Fatal("expected old program to be deactivated")
	}
	if len(santriProgramRepo.saved) != 1 {
		t.Fatalf("expected 1 new santri program saved, got %d", len(santriProgramRepo.saved))
	}
}

func TestAssignSantriProgramAdminSameProgramIdempotent(t *testing.T) {
	prog := &progEntity.Program{ID: "prog-1", Code: "TAHFIDZ", Name: "Tahfidz", Status: "active"}
	programRepo := &fakeProgramRepo{programByID: prog}

	santriProgramRepo := &fakeSantriProgramRepo{
		activeBySantri: &spEntity.SantriProgram{ID: "sp-1", SantriID: "santri-1", ProgramID: "prog-1", IsActive: true},
	}

	uc := NewAssignSantriProgramAdminUseCase(santriProgramRepo, programRepo, fakeTransactor{})
	resp, err := uc.Execute(context.Background(), "santri-1", "prog-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ProgramID != "prog-1" {
		t.Fatalf("expected program prog-1, got %s", resp.ProgramID)
	}
	if santriProgramRepo.deactivated {
		t.Fatal("expected no deactivation for same program")
	}
	if len(santriProgramRepo.saved) != 0 {
		t.Fatalf("expected no new record for same program, got %d", len(santriProgramRepo.saved))
	}
}

func TestAssignSantriProgramAdminInactiveProgram(t *testing.T) {
	prog := &progEntity.Program{ID: "prog-2", Code: "KITAB", Name: "Kitab", Status: "inactive"}
	programRepo := &fakeProgramRepo{programByID: prog}
	santriProgramRepo := &fakeSantriProgramRepo{}

	uc := NewAssignSantriProgramAdminUseCase(santriProgramRepo, programRepo, fakeTransactor{})
	_, err := uc.Execute(context.Background(), "santri-1", "prog-2")
	if err == nil {
		t.Fatal("expected error for inactive program")
	}
}

// --- Santri request transfer ---

func TestRequestProgramTransferSuccess(t *testing.T) {
	current := &spEntity.SantriProgram{ID: "sp-1", SantriID: "santri-1", ProgramID: "prog-1", IsActive: true}
	toProg := &progEntity.Program{ID: "prog-2", Code: "KITAB", Name: "Kitab", Status: "active"}
	fromProg := &progEntity.Program{ID: "prog-1", Code: "TAHFIDZ", Name: "Tahfidz", Status: "active"}

	uc := NewRequestProgramTransferUseCase(
		&fakePtrRepo{},
		&fakeSantriProgramRepo{activeBySantri: current},
		&fakeProgramRepo{programByID: toProg, programByID2: fromProg},
		&fakeKesantrian{byUserID: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI"}},
	)

	resp, err := uc.Execute(context.Background(), "user-1", dto.RequestProgramTransferRequest{ToProgramID: "prog-2", Notes: strP("pindah")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "pending" {
		t.Fatalf("expected pending, got %s", resp.Status)
	}
	if resp.FromProgramID != "prog-1" || resp.ToProgramID != "prog-2" {
		t.Fatalf("unexpected programs: from=%s to=%s", resp.FromProgramID, resp.ToProgramID)
	}
}

func TestRequestProgramTransferRejectsSameProgram(t *testing.T) {
	current := &spEntity.SantriProgram{ID: "sp-1", SantriID: "santri-1", ProgramID: "prog-1", IsActive: true}
	toProg := &progEntity.Program{ID: "prog-1", Code: "TAHFIDZ", Name: "Tahfidz", Status: "active"}

	uc := NewRequestProgramTransferUseCase(
		&fakePtrRepo{},
		&fakeSantriProgramRepo{activeBySantri: current},
		&fakeProgramRepo{programByID: toProg},
		&fakeKesantrian{byUserID: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI"}},
	)

	_, err := uc.Execute(context.Background(), "user-1", dto.RequestProgramTransferRequest{ToProgramID: "prog-1"})
	if err == nil {
		t.Fatal("expected error for same program transfer")
	}
}

func TestRequestProgramTransferRejectsExistingPending(t *testing.T) {
	current := &spEntity.SantriProgram{ID: "sp-1", SantriID: "santri-1", ProgramID: "prog-1", IsActive: true}
	toProg := &progEntity.Program{ID: "prog-2", Code: "KITAB", Name: "Kitab", Status: "active"}

	uc := NewRequestProgramTransferUseCase(
		&fakePtrRepo{pending: newPtrFixture("req-1", "santri-1", "prog-1", "prog-2")},
		&fakeSantriProgramRepo{activeBySantri: current},
		&fakeProgramRepo{programByID: toProg},
		&fakeKesantrian{byUserID: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI"}},
	)

	_, err := uc.Execute(context.Background(), "user-1", dto.RequestProgramTransferRequest{ToProgramID: "prog-2"})
	if err == nil {
		t.Fatal("expected conflict error for existing pending request")
	}
}

// --- Admin approve ---

func TestApproveProgramTransfer(t *testing.T) {
	transfer := newPtrFixture("req-1", "santri-1", "prog-1", "prog-2")
	fromProg := &progEntity.Program{ID: "prog-1", Code: "TAHFIDZ", Name: "Tahfidz", Status: "active"}
	toProg := &progEntity.Program{ID: "prog-2", Code: "KITAB", Name: "Kitab", Status: "active"}

	ptrRepo := &fakePtrRepo{byID: transfer}
	santriProgramRepo := &fakeSantriProgramRepo{}

	uc := NewApproveProgramTransferUseCase(ptrRepo, santriProgramRepo, &fakeProgramRepo{programByID: toProg, programByID2: fromProg}, fakeTransactor{})
	resp, err := uc.Execute(context.Background(), "req-1", "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "approved" {
		t.Fatalf("expected approved, got %s", resp.Status)
	}
	if !santriProgramRepo.deactivated {
		t.Fatal("expected old program deactivated on approve")
	}
	if len(santriProgramRepo.saved) != 1 {
		t.Fatalf("expected 1 new santri program, got %d", len(santriProgramRepo.saved))
	}
	if santriProgramRepo.saved[0].ProgramID != "prog-2" {
		t.Fatalf("expected new program prog-2, got %s", santriProgramRepo.saved[0].ProgramID)
	}
}

func TestApproveProgramTransferAlreadyProcessed(t *testing.T) {
	transfer := newPtrFixture("req-1", "santri-1", "prog-1", "prog-2")
	transfer.Status = "approved"

	uc := NewApproveProgramTransferUseCase(
		&fakePtrRepo{byID: transfer},
		&fakeSantriProgramRepo{},
		&fakeProgramRepo{},
		fakeTransactor{},
	)
	_, err := uc.Execute(context.Background(), "req-1", "admin-1")
	if err == nil {
		t.Fatal("expected error for already approved request")
	}
}

// --- Admin reject ---

func TestRejectProgramTransfer(t *testing.T) {
	transfer := newPtrFixture("req-1", "santri-1", "prog-1", "prog-2")
	ptrRepo := &fakePtrRepo{byID: transfer}

	uc := NewRejectProgramTransferUseCase(ptrRepo, &fakeProgramRepo{})
	resp, err := uc.Execute(context.Background(), "req-1", "admin-1", dto.RejectProgramTransferRequest{AdminNotes: strP("tidak ada kuota")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "rejected" {
		t.Fatalf("expected rejected, got %s", resp.Status)
	}
	if resp.AdminNotes == nil || *resp.AdminNotes != "tidak ada kuota" {
		t.Fatalf("expected admin notes, got %v", resp.AdminNotes)
	}
}

func TestRejectProgramTransferAlreadyProcessed(t *testing.T) {
	transfer := newPtrFixture("req-1", "santri-1", "prog-1", "prog-2")
	transfer.Status = "rejected"

	uc := NewRejectProgramTransferUseCase(&fakePtrRepo{byID: transfer}, &fakeProgramRepo{})
	_, err := uc.Execute(context.Background(), "req-1", "admin-1", dto.RejectProgramTransferRequest{})
	if err == nil {
		t.Fatal("expected error for already rejected request")
	}
}
