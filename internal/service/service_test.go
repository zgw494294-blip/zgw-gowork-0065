package service

import (
	"context"
	"path/filepath"
	"testing"

	"maskreview/internal/domain"
	"maskreview/internal/store"
)

func newTestService(t *testing.T) (*Service, *store.FileStore) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")
	st, err := store.NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	return NewService(st), st
}

func TestNormalChain(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	// 创建产品层级
	pl, err := svc.CreateProductLevel(ctx, "Layer1")
	if err != nil {
		t.Fatal(err)
	}
	// 创建版本
	mv, err := svc.CreateMaskVersion(ctx, pl.ID, 1, "initial")
	if err != nil {
		t.Fatal(err)
	}
	// 创建变更申请
	cr, err := svc.CreateChangeRequest(ctx, mv.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 送审
	if err := svc.SubmitChangeRequest(ctx, cr.ID); err != nil {
		t.Fatal(err)
	}
	// 创建验证批次
	batch, err := svc.CreateVerificationBatch(ctx, cr.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 提交验证通过
	if err := svc.SetVerificationResult(ctx, batch.ID, true, "all passed"); err != nil {
		t.Fatal(err)
	}
	// 审核通过
	if err := svc.AuditChangeRequest(ctx, cr.ID, domain.ConclusionPass, "ok"); err != nil {
		t.Fatal(err)
	}
	// 启用
	if err := svc.EnableMaskVersion(ctx, mv.ID); err != nil {
		t.Fatal(err)
	}

	// 检查状态
	snap := svc.store.GetSnapshot()
	if snap.MaskVersions[0].Status != domain.StatusActive {
		t.Fatalf("expected active, got %s", snap.MaskVersions[0].Status)
	}
}

func TestFailureAfterStateChange(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	pl, _ := svc.CreateProductLevel(ctx, "L")
	mv, _ := svc.CreateMaskVersion(ctx, pl.ID, 1, "")
	cr, _ := svc.CreateChangeRequest(ctx, mv.ID)
	_ = svc.SubmitChangeRequest(ctx, cr.ID)
	batch, _ := svc.CreateVerificationBatch(ctx, cr.ID)
	// 验证失败
	if err := svc.SetVerificationResult(ctx, batch.ID, false, "critical fail"); err != nil {
		t.Fatal(err)
	}
	// 检查状态不变
	snap := svc.store.GetSnapshot()
	if snap.MaskVersions[0].Status != domain.StatusSubmitted {
		t.Fatalf("expected submitted after failure, got %s", snap.MaskVersions[0].Status)
	}
	// 尝试审核（应该失败，因为状态不是verified）
	if err := svc.AuditChangeRequest(ctx, cr.ID, domain.ConclusionPass, "should fail"); err == nil {
		t.Fatal("expected audit to fail")
	} else if err.Error() != "change request must be verified" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSingleActiveVersionConstraint(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	pl, _ := svc.CreateProductLevel(ctx, "L")
	mv1, _ := svc.CreateMaskVersion(ctx, pl.ID, 1, "")
	cr1, _ := svc.CreateChangeRequest(ctx, mv1.ID)
	_ = svc.SubmitChangeRequest(ctx, cr1.ID)
	b1, _ := svc.CreateVerificationBatch(ctx, cr1.ID)
	_ = svc.SetVerificationResult(ctx, b1.ID, true, "")
	_ = svc.AuditChangeRequest(ctx, cr1.ID, domain.ConclusionPass, "")
	_ = svc.EnableMaskVersion(ctx, mv1.ID)

	mv2, _ := svc.CreateMaskVersion(ctx, pl.ID, 2, "new")
	cr2, _ := svc.CreateChangeRequest(ctx, mv2.ID)
	_ = svc.SubmitChangeRequest(ctx, cr2.ID)
	b2, _ := svc.CreateVerificationBatch(ctx, cr2.ID)
	_ = svc.SetVerificationResult(ctx, b2.ID, true, "")
	_ = svc.AuditChangeRequest(ctx, cr2.ID, domain.ConclusionPass, "")
	_ = svc.EnableMaskVersion(ctx, mv2.ID)

	snap := svc.store.GetSnapshot()
	for _, mv := range snap.MaskVersions {
		if mv.ID == mv1.ID && mv.Status != domain.StatusSuperseded {
			t.Fatalf("mv1 should be superseded, got %s", mv.Status)
		}
		if mv.ID == mv2.ID && mv.Status != domain.StatusActive {
			t.Fatalf("mv2 should be active, got %s", mv.Status)
		}
	}
}

func TestRedraftInvalidatesAudit(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	pl, _ := svc.CreateProductLevel(ctx, "L")
	mv, _ := svc.CreateMaskVersion(ctx, pl.ID, 1, "")
	cr, _ := svc.CreateChangeRequest(ctx, mv.ID)
	_ = svc.SubmitChangeRequest(ctx, cr.ID)
	b, _ := svc.CreateVerificationBatch(ctx, cr.ID)
	_ = svc.SetVerificationResult(ctx, b.ID, true, "")
	_ = svc.AuditChangeRequest(ctx, cr.ID, domain.ConclusionPass, "ok")

	// 重新起草
	if err := svc.RedraftMaskVersion(ctx, mv.ID); err != nil {
		t.Fatal(err)
	}
	snap := svc.store.GetSnapshot()
	if snap.MaskVersions[0].Status != domain.StatusDraft {
		t.Fatalf("expected draft, got %s", snap.MaskVersions[0].Status)
	}
	if snap.ChangeRequests[0].AuditResult != "" {
		t.Fatalf("audit should be invalidated, got %s", snap.ChangeRequests[0].AuditResult)
	}
}

func TestDerivedDiffSummary(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	pl, _ := svc.CreateProductLevel(ctx, "L")
	mv1, _ := svc.CreateMaskVersion(ctx, pl.ID, 1, "old")
	mv2, _ := svc.CreateMaskVersion(ctx, pl.ID, 2, "new")
	summary, err := svc.DiffSummary(ctx, pl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary["diff_count"].(int) != 2 {
		t.Fatalf("expected 2 versions, got %v", summary["diff_count"])
	}
	_ = mv1
	_ = mv2
}

func TestFailedVerificationCannotBeAudited(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	pl, err := svc.CreateProductLevel(ctx, "failed-verification")
	if err != nil {
		t.Fatal(err)
	}
	mv, err := svc.CreateMaskVersion(ctx, pl.ID, 1, "trial")
	if err != nil {
		t.Fatal(err)
	}
	cr, err := svc.CreateChangeRequest(ctx, mv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.SubmitChangeRequest(ctx, cr.ID); err != nil {
		t.Fatal(err)
	}
	batch, err := svc.CreateVerificationBatch(ctx, cr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.SetVerificationResult(ctx, batch.ID, false, "critical item failed"); err != nil {
		t.Fatal(err)
	}

	if err := svc.AuditChangeRequest(ctx, cr.ID, domain.ConclusionPass, "must not pass"); err == nil {
		t.Fatal("failed verification request was accepted for audit")
	}
}
