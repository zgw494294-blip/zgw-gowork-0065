package domain

import "context"

// 业务服务接口
type MaskReviewService interface {
	CreateProductLevel(ctx context.Context, name string) (*ProductLevel, error)
	ListProductLevels(ctx context.Context) ([]*ProductLevel, error)
	CreateMaskVersion(ctx context.Context, levelID string, version int, description string) (*MaskVersion, error)
	ListMaskVersions(ctx context.Context, levelID string) ([]*MaskVersion, error)
	CreateChangeRequest(ctx context.Context, maskVersionID string) (*ChangeRequest, error)
	ListChangeRequests(ctx context.Context) ([]*ChangeRequest, error)
	SubmitChangeRequest(ctx context.Context, crID string) error
	CreateVerificationBatch(ctx context.Context, crID string) (*VerificationBatch, error)
	SetVerificationResult(ctx context.Context, batchID string, criticalPassed bool, note string) error
	AuditChangeRequest(ctx context.Context, crID string, conclusion AuditConclusion, reason string) error
	EnableMaskVersion(ctx context.Context, mvID string) error
	RedraftMaskVersion(ctx context.Context, mvID string) error
	DiffSummary(ctx context.Context, levelID string) (map[string]interface{}, error)
	ExportPendingAudit(ctx context.Context) ([]*ChangeRequest, error)
}
