package store

import "maskreview/internal/domain"

func cloneSnapshot(s Snapshot) Snapshot {
	c := Snapshot{
		ProductLevels:       make([]*domain.ProductLevel, len(s.ProductLevels)),
		MaskVersions:        make([]*domain.MaskVersion, len(s.MaskVersions)),
		ChangeRequests:      make([]*domain.ChangeRequest, len(s.ChangeRequests)),
		VerificationBatches: make([]*domain.VerificationBatch, len(s.VerificationBatches)),
		AuditRecords:        make([]*domain.AuditConclusionRecord, len(s.AuditRecords)),
	}
	for i, v := range s.ProductLevels {
		vv := *v
		c.ProductLevels[i] = &vv
	}
	for i, v := range s.MaskVersions {
		vv := *v
		c.MaskVersions[i] = &vv
	}
	for i, v := range s.ChangeRequests {
		vv := *v
		c.ChangeRequests[i] = &vv
	}
	for i, v := range s.VerificationBatches {
		vv := *v
		c.VerificationBatches[i] = &vv
	}
	for i, v := range s.AuditRecords {
		vv := *v
		c.AuditRecords[i] = &vv
	}
	return c
}
