package httpapi

import (
	"maskreview/internal/service"
)

func NewHandler(svc *service.Service) *Router {
	h := &Handler{svc: svc}
	r := NewRouter()
	r.HandleFunc("POST", PathProductLevels, h.CreateProductLevel)
	r.HandleFunc("GET", PathProductLevels, h.ListProductLevels)
	r.HandleFunc("POST", PathMaskVersions, h.CreateMaskVersion)
	r.HandleFunc("GET", PathMaskVersions, h.ListMaskVersions)
	r.HandleFunc("POST", PathChangeRequests, h.CreateChangeRequest)
	r.HandleFunc("GET", PathChangeRequests, h.ListChangeRequests)
	r.HandleFunc("POST", PathSubmitChangeRequest, h.SubmitChangeRequest)
	r.HandleFunc("POST", PathVerificationBatches, h.CreateVerificationBatch)
	r.HandleFunc("POST", PathVerificationResult, h.SetVerificationResult)
	r.HandleFunc("POST", PathAuditChangeRequest, h.AuditChangeRequest)
	r.HandleFunc("POST", PathEnableMaskVersion, h.EnableMaskVersion)
	r.HandleFunc("POST", PathRedraftMaskVersion, h.RedraftMaskVersion)
	r.HandleFunc("GET", PathDiffSummary, h.DiffSummary)
	r.HandleFunc("GET", PathExportPendingAudit, h.ExportPendingAudit)
	return r
}
