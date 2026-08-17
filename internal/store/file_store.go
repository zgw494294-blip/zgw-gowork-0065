package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"maskreview/internal/domain"
)

// 数据快照
type Snapshot struct {
	ProductLevels       []*domain.ProductLevel       `json:"product_levels"`
	MaskVersions        []*domain.MaskVersion        `json:"mask_versions"`
	ChangeRequests      []*domain.ChangeRequest      `json:"change_requests"`
	VerificationBatches []*domain.VerificationBatch  `json:"verification_batches"`
	AuditRecords        []*domain.AuditConclusionRecord `json:"audit_records"`
}

type FileStore struct {
	mu      sync.RWMutex
	path    string
	snapshot Snapshot
}

func NewFileStore(path string) (*FileStore, error) {
	s := &FileStore{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.snapshot = Snapshot{
			ProductLevels:       make([]*domain.ProductLevel, 0),
			MaskVersions:        make([]*domain.MaskVersion, 0),
			ChangeRequests:      make([]*domain.ChangeRequest, 0),
			VerificationBatches: make([]*domain.VerificationBatch, 0),
			AuditRecords:        make([]*domain.AuditConclusionRecord, 0),
		}
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.snapshot)
}

func (s *FileStore) SaveSnapshot(snap Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = snap
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *FileStore) GetSnapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneSnapshot(s.snapshot)
}
