package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"maskreview/internal/domain"
	"maskreview/internal/store"
)

type Service struct {
	store *store.FileStore
	mu    sync.Mutex
}

func NewService(st *store.FileStore) *Service {
	return &Service{store: st}
}

func (s *Service) CreateProductLevel(ctx context.Context, name string) (*domain.ProductLevel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	for _, pl := range snap.ProductLevels {
		if pl.Name == name {
			return nil, errors.New("product level name already exists")
		}
	}
	pl := &domain.ProductLevel{
		ID:   fmt.Sprintf("pl_%d", time.Now().UnixNano()),
		Name: name,
	}
	snap.ProductLevels = append(snap.ProductLevels, pl)
	if err := s.store.SaveSnapshot(snap); err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *Service) ListProductLevels(ctx context.Context) ([]*domain.ProductLevel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.GetSnapshot().ProductLevels, nil
}

func (s *Service) CreateMaskVersion(ctx context.Context, levelID string, version int, description string) (*domain.MaskVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var level *domain.ProductLevel
	for _, pl := range snap.ProductLevels {
		if pl.ID == levelID {
			level = pl
			break
		}
	}
	if level == nil {
		return nil, errors.New("product level not found")
	}
	for _, mv := range snap.MaskVersions {
		if mv.LevelID == levelID && mv.Version == version {
			return nil, errors.New("version already exists for this level")
		}
	}
	mv := &domain.MaskVersion{
		ID:          fmt.Sprintf("mv_%d", time.Now().UnixNano()),
		LevelID:     levelID,
		Version:     version,
		Description: description,
		Status:      domain.StatusDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	snap.MaskVersions = append(snap.MaskVersions, mv)
	if err := s.store.SaveSnapshot(snap); err != nil {
		return nil, err
	}
	return mv, nil
}

func (s *Service) ListMaskVersions(ctx context.Context, levelID string) ([]*domain.MaskVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	all := snap.MaskVersions
	if levelID == "" {
		return all, nil
	}
	filtered := make([]*domain.MaskVersion, 0)
	for _, mv := range all {
		if mv.LevelID == levelID {
			filtered = append(filtered, mv)
		}
	}
	return filtered, nil
}

func (s *Service) CreateChangeRequest(ctx context.Context, maskVersionID string) (*domain.ChangeRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var mv *domain.MaskVersion
	for _, m := range snap.MaskVersions {
		if m.ID == maskVersionID {
			mv = m
			break
		}
	}
	if mv == nil {
		return nil, errors.New("mask version not found")
	}
	if mv.Status != domain.StatusDraft && mv.Status != domain.StatusSuperseded {
		return nil, errors.New("mask version must be in draft or superseded")
	}
	cr := &domain.ChangeRequest{
		ID:            fmt.Sprintf("cr_%d", time.Now().UnixNano()),
		MaskVersionID: maskVersionID,
		Status:        domain.CRDraft,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	snap.ChangeRequests = append(snap.ChangeRequests, cr)
	if err := s.store.SaveSnapshot(snap); err != nil {
		return nil, err
	}
	return cr, nil
}

func (s *Service) ListChangeRequests(ctx context.Context) ([]*domain.ChangeRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.GetSnapshot().ChangeRequests, nil
}

func (s *Service) SubmitChangeRequest(ctx context.Context, crID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var cr *domain.ChangeRequest
	for _, c := range snap.ChangeRequests {
		if c.ID == crID {
			cr = c
			break
		}
	}
	if cr == nil {
		return errors.New("change request not found")
	}
	if cr.Status != domain.CRDraft {
		return errors.New("only draft can be submitted")
	}
	var mv *domain.MaskVersion
	for _, m := range snap.MaskVersions {
		if m.ID == cr.MaskVersionID {
			mv = m
			break
		}
	}
	if mv == nil {
		return errors.New("mask version not found")
	}
	if mv.Status != domain.StatusDraft {
		return errors.New("mask version must be draft")
	}
	cr.Status = domain.CRSubmitted
	mv.Status = domain.StatusSubmitted
	cr.UpdatedAt = time.Now()
	mv.UpdatedAt = time.Now()
	if err := s.store.SaveSnapshot(snap); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateVerificationBatch(ctx context.Context, crID string) (*domain.VerificationBatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var cr *domain.ChangeRequest
	for _, c := range snap.ChangeRequests {
		if c.ID == crID {
			cr = c
			break
		}
	}
	if cr == nil {
		return nil, errors.New("change request not found")
	}
	if cr.Status != domain.CRSubmitted && cr.Status != domain.CRVerified {
		return nil, errors.New("change request must be submitted or verified")
	}
	var mv *domain.MaskVersion
	for _, m := range snap.MaskVersions {
		if m.ID == cr.MaskVersionID {
			mv = m
			break
		}
	}
	if mv == nil {
		return nil, errors.New("mask version not found")
	}
	if mv.Status != domain.StatusSubmitted && mv.Status != domain.StatusInTrial {
		return nil, errors.New("mask version must be submitted or in_trial")
	}
	// 已经存在成功批次的不能重复创建
	for _, b := range snap.VerificationBatches {
		if b.ChangeRequestID == crID && b.Status == "passed" {
			return nil, errors.New("batch already passed")
		}
	}
	batch := &domain.VerificationBatch{
		ID:              fmt.Sprintf("vb_%d", time.Now().UnixNano()),
		ChangeRequestID: crID,
		MaskVersionID:   mv.ID,
		Status:          "pending",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	snap.VerificationBatches = append(snap.VerificationBatches, batch)
	mv.Status = domain.StatusInTrial
	cr.Status = domain.CRVerified // 实际上试投阶段
	mv.UpdatedAt = time.Now()
	cr.UpdatedAt = time.Now()
	if err := s.store.SaveSnapshot(snap); err != nil {
		return nil, err
	}
	return batch, nil
}

func (s *Service) SetVerificationResult(ctx context.Context, batchID string, criticalPassed bool, note string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var batch *domain.VerificationBatch
	for _, b := range snap.VerificationBatches {
		if b.ID == batchID {
			batch = b
			break
		}
	}
	if batch == nil {
		return errors.New("batch not found")
	}
	if batch.Status != "pending" {
		return errors.New("batch already processed")
	}
	var cr *domain.ChangeRequest
	for _, c := range snap.ChangeRequests {
		if c.ID == batch.ChangeRequestID {
			cr = c
			break
		}
	}
	if cr == nil {
		return errors.New("change request not found")
	}
	var mv *domain.MaskVersion
	for _, m := range snap.MaskVersions {
		if m.ID == batch.MaskVersionID {
			mv = m
			break
		}
	}
	if mv == nil {
		return errors.New("mask version not found")
	}
	batch.CriticalPassed = criticalPassed
	batch.Note = note
	batch.UpdatedAt = time.Now()
	if criticalPassed {
		batch.Status = "passed"
		cr.Status = domain.CRVerified
		mv.Status = domain.StatusVerified
	} else {
		batch.Status = "failed"
		cr.Status = domain.CRSubmitted
		mv.Status = domain.StatusSubmitted // 回到送审，但保持记录
	}
	cr.UpdatedAt = time.Now()
	mv.UpdatedAt = time.Now()
	if err := s.store.SaveSnapshot(snap); err != nil {
		return err
	}
	return nil
}

func (s *Service) AuditChangeRequest(ctx context.Context, crID string, conclusion domain.AuditConclusion, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var cr *domain.ChangeRequest
	for _, c := range snap.ChangeRequests {
		if c.ID == crID {
			cr = c
			break
		}
	}
	if cr == nil {
		return errors.New("change request not found")
	}
	if cr.Status != domain.CRVerified {
		return errors.New("change request must be verified")
	}
	var mv *domain.MaskVersion
	for _, m := range snap.MaskVersions {
		if m.ID == cr.MaskVersionID {
			mv = m
			break
		}
	}
	if mv == nil {
		return errors.New("mask version not found")
	}
	record := &domain.AuditConclusionRecord{
		ID:              fmt.Sprintf("ac_%d", time.Now().UnixNano()),
		ChangeRequestID: crID,
		MaskVersionID:   mv.ID,
		Conclusion:      conclusion,
		Reason:          reason,
		CreatedAt:       time.Now(),
	}
	snap.AuditRecords = append(snap.AuditRecords, record)
	cr.AuditResult = conclusion
	cr.AuditReason = reason
	cr.UpdatedAt = time.Now()
	cr.Status = domain.CRAudited
	switch conclusion {
	case domain.ConclusionPass:
		mv.Status = domain.StatusVerified // 保持已验证，等待启用
	case domain.ConclusionFail:
		mv.Status = domain.StatusSuperseded
	case domain.ConclusionRedraft:
		mv.Status = domain.StatusDraft
	default:
		if err := s.store.SaveSnapshot(snap); err != nil {
			return err
		}
		return errors.New("invalid conclusion")
	}
	mv.UpdatedAt = time.Now()
	if err := s.store.SaveSnapshot(snap); err != nil {
		return err
	}
	return nil
}

func (s *Service) EnableMaskVersion(ctx context.Context, mvID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var mv *domain.MaskVersion
	for _, m := range snap.MaskVersions {
		if m.ID == mvID {
			mv = m
			break
		}
	}
	if mv == nil {
		return errors.New("mask version not found")
	}
	if mv.Status != domain.StatusVerified {
		return errors.New("mask version must be verified")
	}
	// 检查该level下是否已有启用版本
	for _, other := range snap.MaskVersions {
		if other.LevelID == mv.LevelID && other.Status == domain.StatusActive && other.ID != mv.ID {
			// 使旧版本失效
			other.Status = domain.StatusSuperseded
			other.SupersededBy = mv.ID
			other.UpdatedAt = time.Now()
		}
	}
	mv.Status = domain.StatusActive
	mv.UpdatedAt = time.Now()
	if err := s.store.SaveSnapshot(snap); err != nil {
		return err
	}
	return nil
}

func (s *Service) RedraftMaskVersion(ctx context.Context, mvID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	var mv *domain.MaskVersion
	for _, m := range snap.MaskVersions {
		if m.ID == mvID {
			mv = m
			break
		}
	}
	if mv == nil {
		return errors.New("mask version not found")
	}
	if mv.Status != domain.StatusVerified && mv.Status != domain.StatusActive && mv.Status != domain.StatusSuperseded {
		return errors.New("cannot redraft in current status")
	}
	// 使所有相关变更申请的审核结论失效
	for _, cr := range snap.ChangeRequests {
		if cr.MaskVersionID == mvID && cr.Status == domain.CRAudited {
			cr.AuditResult = ""
			cr.AuditReason = "redrafted"
			cr.UpdatedAt = time.Now()
		}
	}
	mv.Status = domain.StatusDraft
	mv.UpdatedAt = time.Now()
	if err := s.store.SaveSnapshot(snap); err != nil {
		return err
	}
	return nil
}

func (s *Service) DiffSummary(ctx context.Context, levelID string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	versions := make([]*domain.MaskVersion, 0)
	for _, mv := range snap.MaskVersions {
		if mv.LevelID == levelID {
			versions = append(versions, mv)
		}
	}
	if len(versions) == 0 {
		return map[string]interface{}{"level_id": levelID, "versions": []interface{}{}}, nil
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].Version < versions[j].Version })
	type versionItem struct {
		ID          string `json:"id"`
		Version     int    `json:"version"`
		Status      string `json:"status"`
		Description string `json:"description"`
	}
	items := make([]versionItem, 0, len(versions))
	for _, mv := range versions {
		items = append(items, versionItem{ID: mv.ID, Version: mv.Version, Status: string(mv.Status), Description: mv.Description})
	}
	return map[string]interface{}{
		"level_id":   levelID,
		"versions":   items,
		"diff_count": len(items),
	}, nil
}

func (s *Service) ExportPendingAudit(ctx context.Context) ([]*domain.ChangeRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.store.GetSnapshot()
	pending := make([]*domain.ChangeRequest, 0)
	for _, cr := range snap.ChangeRequests {
		if cr.Status == domain.CRVerified {
			pending = append(pending, cr)
		}
	}
	return pending, nil
}

// 辅助：将待审核清单导出为CSV（也可作为非API能力）
func (s *Service) ExportPendingAuditCSV(ctx context.Context, w io.Writer) error {
	pending, err := s.ExportPendingAudit(ctx)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"ID", "MaskVersionID", "Status", "CreatedAt"}); err != nil {
		return err
	}
	for _, cr := range pending {
		if err := cw.Write([]string{cr.ID, cr.MaskVersionID, string(cr.Status), cr.CreatedAt.Format(time.RFC3339)}); err != nil {
			return err
		}
	}
	return nil
}

// 确保 Service 实现 domain.MaskReviewService
var _ domain.MaskReviewService = (*Service)(nil)
