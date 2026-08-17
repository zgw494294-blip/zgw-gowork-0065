package httpapi

const (
	PathProductLevels         = "/api/product-levels"
	PathMaskVersions          = "/api/mask-versions"
	PathChangeRequests        = "/api/change-requests"
	PathSubmitChangeRequest   = "/api/change-requests/{id}/submit"
	PathVerificationBatches   = "/api/verification-batches"
	PathVerificationResult    = "/api/verification-batches/{id}/result"
	PathAuditChangeRequest    = "/api/change-requests/{id}/audit"
	PathEnableMaskVersion     = "/api/mask-versions/{id}/enable"
	PathRedraftMaskVersion    = "/api/mask-versions/{id}/redraft"
	PathDiffSummary           = "/api/summary/diff"
	PathExportPendingAudit    = "/api/export/pending-audit"
)
