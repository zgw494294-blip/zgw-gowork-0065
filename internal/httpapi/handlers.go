package httpapi

import (
	"encoding/json"
	"net/http"

	"maskreview/internal/domain"
	"maskreview/internal/service"
)

type Handler struct {
	svc *service.Service
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// --- Product Levels ---
func (h *Handler) CreateProductLevel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	pl, err := h.svc.CreateProductLevel(r.Context(), req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, pl)
}

func (h *Handler) ListProductLevels(w http.ResponseWriter, r *http.Request) {
	pls, err := h.svc.ListProductLevels(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pls)
}

// --- Mask Versions ---
func (h *Handler) CreateMaskVersion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LevelID     string `json:"level_id"`
		Version     int    `json:"version"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	mv, err := h.svc.CreateMaskVersion(r.Context(), req.LevelID, req.Version, req.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, mv)
}

func (h *Handler) ListMaskVersions(w http.ResponseWriter, r *http.Request) {
	levelID := r.URL.Query().Get("level_id")
	mvs, err := h.svc.ListMaskVersions(r.Context(), levelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mvs)
}

// --- Change Requests ---
func (h *Handler) CreateChangeRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MaskVersionID string `json:"mask_version_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	cr, err := h.svc.CreateChangeRequest(r.Context(), req.MaskVersionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cr)
}

func (h *Handler) ListChangeRequests(w http.ResponseWriter, r *http.Request) {
	rs, err := h.svc.ListChangeRequests(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rs)
}

func (h *Handler) SubmitChangeRequest(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	if err := h.svc.SubmitChangeRequest(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}

// --- Verification Batches ---
func (h *Handler) CreateVerificationBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChangeRequestID string `json:"change_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	vb, err := h.svc.CreateVerificationBatch(r.Context(), req.ChangeRequestID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, vb)
}

func (h *Handler) SetVerificationResult(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	var req struct {
		CriticalPassed bool   `json:"critical_passed"`
		Note           string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.SetVerificationResult(r.Context(), id, req.CriticalPassed, req.Note); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Audit ---
func (h *Handler) AuditChangeRequest(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	var req struct {
		Conclusion string `json:"conclusion"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	conclusion := domain.AuditConclusion(req.Conclusion)
	if err := h.svc.AuditChangeRequest(r.Context(), id, conclusion, req.Reason); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "audited"})
}

// --- Enable / Redraft ---
func (h *Handler) EnableMaskVersion(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	if err := h.svc.EnableMaskVersion(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

func (h *Handler) RedraftMaskVersion(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	if err := h.svc.RedraftMaskVersion(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "drafted"})
}

// --- Summary and Export ---
func (h *Handler) DiffSummary(w http.ResponseWriter, r *http.Request) {
	levelID := r.URL.Query().Get("level")
	if levelID == "" {
		writeError(w, http.StatusBadRequest, "missing level")
		return
	}
	summary, err := h.svc.DiffSummary(r.Context(), levelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) ExportPendingAudit(w http.ResponseWriter, r *http.Request) {
	pending, err := h.svc.ExportPendingAudit(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var resp []map[string]interface{}
	for _, cr := range pending {
		resp = append(resp, map[string]interface{}{
			"id":              cr.ID,
			"mask_version_id": cr.MaskVersionID,
			"status":          cr.Status,
			"created_at":      cr.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
