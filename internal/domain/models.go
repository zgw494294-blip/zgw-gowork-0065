package domain

import "time"

// 状态常量
type MaskVersionStatus string

const (
	StatusDraft      MaskVersionStatus = "draft"
	StatusSubmitted  MaskVersionStatus = "submitted"
	StatusInTrial    MaskVersionStatus = "in_trial"
	StatusVerified   MaskVersionStatus = "verified"
	StatusActive     MaskVersionStatus = "active"
	StatusSuperseded MaskVersionStatus = "superseded"
)

type ChangeRequestStatus string

const (
	CRDraft     ChangeRequestStatus = "draft"
	CRSubmitted ChangeRequestStatus = "submitted"
	CRVerified  ChangeRequestStatus = "verified"
	CRAudited   ChangeRequestStatus = "audited"
)

type AuditConclusion string

const (
	ConclusionPass    AuditConclusion = "pass"
	ConclusionFail    AuditConclusion = "fail"
	ConclusionRedraft AuditConclusion = "redraft"
)

type ProductLevel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MaskVersion struct {
	ID           string            `json:"id"`
	LevelID      string            `json:"level_id"`
	Version      int               `json:"version"`
	Description  string            `json:"description"`
	Status       MaskVersionStatus `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	SupersededBy string            `json:"superseded_by,omitempty"`
}

type ChangeRequest struct {
	ID            string              `json:"id"`
	MaskVersionID string              `json:"mask_version_id"`
	Status        ChangeRequestStatus `json:"status"`
	AuditResult   AuditConclusion     `json:"audit_result,omitempty"`
	AuditReason   string              `json:"audit_reason,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type VerificationBatch struct {
	ID              string `json:"id"`
	ChangeRequestID string `json:"change_request_id"`
	MaskVersionID   string `json:"mask_version_id"`
	Status          string `json:"status"` // pending, passed, failed
	CriticalPassed  bool   `json:"critical_passed"`
	Note            string `json:"note"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AuditConclusionRecord struct {
	ID              string          `json:"id"`
	ChangeRequestID string          `json:"change_request_id"`
	MaskVersionID   string          `json:"mask_version_id"`
	Conclusion      AuditConclusion `json:"conclusion"`
	Reason          string          `json:"reason"`
	CreatedAt       time.Time       `json:"created_at"`
}
