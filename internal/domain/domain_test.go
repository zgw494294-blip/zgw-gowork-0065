package domain

import (
	"testing"
)

func TestValidateBatchMaskVersion(t *testing.T) {
	cr := &ChangeRequest{ID: "cr1", MaskVersionID: "mv1", Status: CRSubmitted}
	mv := &MaskVersion{ID: "mv1", Status: StatusSubmitted}
	batch := &VerificationBatch{ID: "vb1", ChangeRequestID: "cr1", MaskVersionID: "mv1"}
	if err := ValidateBatchMaskVersion(batch, cr, mv); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// 错误绑定
	batch.MaskVersionID = "mv2"
	if err := ValidateBatchMaskVersion(batch, cr, mv); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateEnable(t *testing.T) {
	mv := &MaskVersion{ID: "mv1", Status: StatusVerified}
	batch := &VerificationBatch{CriticalPassed: true}
	if err := ValidateEnable(mv, batch); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	batch.CriticalPassed = false
	if err := ValidateEnable(mv, batch); err == nil {
		t.Fatal("expected error")
	}
}

func TestFilterMaskVersions(t *testing.T) {
	all := []*MaskVersion{
		{ID: "mv1", LevelID: "L1"},
		{ID: "mv2", LevelID: "L2"},
	}
	filtered := FilterMaskVersions(all, VersionFilter{LevelID: "L1"})
	if len(filtered) != 1 || filtered[0].ID != "mv1" {
		t.Fatalf("unexpected filtered: %v", filtered)
	}
}
