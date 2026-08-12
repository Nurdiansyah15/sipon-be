package entity_test

import (
	"testing"

	"sipon-be/internal/modules/akademik/domain/program_transfer_request/constant"
	"sipon-be/internal/modules/akademik/domain/program_transfer_request/entity"
)

func TestNewProgramTransferRequest(t *testing.T) {
	req, err := entity.NewProgramTransferRequest("req-1", "santri-1", "prog-1", "prog-2", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Status != constant.ProgramTransferRequestStatusPending {
		t.Fatalf("expected pending, got %s", req.Status)
	}
	if req.FromProgramID != "prog-1" || req.ToProgramID != "prog-2" {
		t.Fatalf("unexpected program ids: from=%s to=%s", req.FromProgramID, req.ToProgramID)
	}
}

func TestNewProgramTransferRequestRejectsSameProgram(t *testing.T) {
	_, err := entity.NewProgramTransferRequest("req-1", "santri-1", "prog-1", "prog-1", nil)
	if err == nil {
		t.Fatal("expected error for same program transfer")
	}
}

func TestProgramTransferRequestApprove(t *testing.T) {
	req, _ := entity.NewProgramTransferRequest("req-1", "santri-1", "prog-1", "prog-2", nil)

	if err := req.Approve("admin-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Status != constant.ProgramTransferRequestStatusApproved {
		t.Fatalf("expected approved, got %s", req.Status)
	}
	if req.ReviewedBy == nil || *req.ReviewedBy != "admin-1" {
		t.Fatalf("expected reviewed_by=admin-1, got %v", req.ReviewedBy)
	}
	if req.ReviewedAt == nil {
		t.Fatal("expected reviewed_at to be set")
	}

	// Approve lagi harus gagal.
	if err := req.Approve("admin-2"); err == nil {
		t.Fatal("expected error approving non-pending request")
	}
}

func TestProgramTransferRequestReject(t *testing.T) {
	notes := "tidak ada kuota"
	req, _ := entity.NewProgramTransferRequest("req-1", "santri-1", "prog-1", "prog-2", nil)

	if err := req.Reject("admin-1", &notes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Status != constant.ProgramTransferRequestStatusRejected {
		t.Fatalf("expected rejected, got %s", req.Status)
	}
	if req.AdminNotes == nil || *req.AdminNotes != notes {
		t.Fatalf("expected admin_notes=%q, got %v", notes, req.AdminNotes)
	}
	if req.ReviewedBy == nil || *req.ReviewedBy != "admin-1" {
		t.Fatalf("expected reviewed_by=admin-1, got %v", req.ReviewedBy)
	}

	// Reject lagi harus gagal.
	if err := req.Reject("admin-2", nil); err == nil {
		t.Fatal("expected error rejecting non-pending request")
	}
}
