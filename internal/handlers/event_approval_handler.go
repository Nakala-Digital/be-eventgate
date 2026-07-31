package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"be-eventgate/internal/eventapproval"
	"be-eventgate/internal/httpx"
	"be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
)

// ReviewRequest dipakai untuk approve/reject/request-revision/publish/
// unpublish. Notes WAJIB untuk reject & request-revision (divalidasi
// per-handler), opsional untuk aksi lainnya.
type ReviewRequest struct {
	Notes string `json:"notes"`
}

func parseEventIDParam(r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *EventHandler) loadEvent(id uint) (*models.Event, error) {
	var event models.Event
	if err := h.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// canViewEvent: super_admin/school_reviewer boleh lihat event apa saja;
// admin_panitia hanya boleh lihat event miliknya sendiri. Staf Lapangan
// SENGAJA TIDAK termasuk — scope Staf Lapangan menurut URD hanya scan QR /
// check-in, tidak termasuk melihat detail event maupun approval log.
// Dipakai bersama oleh GetByID (event_handler.go) dan ListApprovalLogs.
func canViewEvent(roleName string, userID uint, event *models.Event) bool {
	if roleName == models.RoleSuperAdmin || roleName == models.RoleSchoolReviewer {
		return true
	}
	if roleName == models.RoleAdminPanitia && event.OrganizerID == userID {
		return true
	}
	return false
}

// ListApprovalLogs mengembalikan seluruh riwayat approval log sebuah event,
// diurutkan dari yang paling lama ke paling baru (berdasarkan log_id, karena
// tabel ini tidak punya kolom created_at).
// GET /api/events/{id}/approval-logs
func (h *EventHandler) ListApprovalLogs(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	event, err := h.loadEvent(id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}

	roleName, _ := middleware.GetRoleName(r.Context())
	userID, _ := middleware.GetUserID(r.Context())
	if !canViewEvent(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to view this event")
		return
	}

	var logs []models.EventApprovalLog
	if err := h.DB.Where("event_id = ?", id).Order("log_id asc").Find(&logs).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load approval logs")
		return
	}

	resp := make([]ApprovalLogResponse, 0, len(logs))
	for _, l := range logs {
		resp = append(resp, toApprovalLogResponse(l))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// SubmitForApproval mengajukan event (yang sebelumnya berstatus draft) untuk direview oleh Super Admin atau Reviewer.
// Logika yang dijalankan:
// 1. Memastikan pengguna yang melakukan submit adalah pemilik event itu sendiri.
// 2. Memvalidasi apakah event memiliki minimal satu jenis tiket; tidak bisa disubmit jika belum ada tiket.
// 3. Mengubah status event menjadi 'pending_approval' dan menambahkan catatan log bahwa event diajukan.
// POST /api/events/{id}/submit (khusus admin_panitia, harus pemilik event)
func (h *EventHandler) SubmitForApproval(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ReviewRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // notes opsional untuk submit

	event, err := h.loadEvent(id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}

	if event.OrganizerID != userID {
		httpx.WriteError(w, http.StatusForbidden, "you can only submit your own event")
		return
	}

	if err := eventapproval.ValidateTransition(event.Status, models.EventStatusPendingApproval); err != nil {
		httpx.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	var ticketTypeCount int64
	if err := h.DB.Model(&models.TicketType{}).Where("event_id = ?", id).Count(&ticketTypeCount).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to validate ticket types")
		return
	}
	if ticketTypeCount == 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "event must have at least one ticket type before submitting for approval")
		return
	}

	now := time.Now()
	newVersion := event.EventVersion + 1

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(event).Updates(map[string]interface{}{
			"status":        models.EventStatusPendingApproval,
			"event_version": newVersion,
			"updated_by":    userID,
		}).Error; err != nil {
			return err
		}

		log := models.EventApprovalLog{
			EventID:       id,
			EventVersion:  newVersion,
			Action:        models.ApprovalActionSubmitted,
			SubmittedByID: &userID,
			SubmittedAt:   &now,
			Notes:         req.Notes,
		}
		return tx.Create(&log).Error
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to submit event for approval")
		return
	}

	event.Status = models.EventStatusPendingApproval
	event.EventVersion = newVersion
	httpx.WriteJSON(w, http.StatusOK, event)
}

// reviewAction memusatkan seluruh logika inti untuk aksi persetujuan (Approve, Reject, Request Revision).
// Alur kerja yang dijalankan pada fungsi ini:
// 1. Memvalidasi transisi status (misalnya, memastikan event sedang berstatus 'pending_approval').
// 2. Memvalidasi catatan wajib (Notes): untuk penolakan dan permintaan revisi, alasan wajib diisi.
// 3. Menegakkan aturan no-self-approval: memastikan reviewer bukanlah penyelenggara event itu sendiri.
// 4. Memperbarui status event dan membuat rekam jejak pada tabel log persetujuan secara transaksional.
func (h *EventHandler) reviewAction(w http.ResponseWriter, r *http.Request, targetStatus, logAction string, notesRequired bool) {
	id, err := parseEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if notesRequired && req.Notes == "" {
		httpx.WriteError(w, http.StatusBadRequest, "notes is required for this action")
		return
	}

	event, err := h.loadEvent(id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}

	// No self-approval: admin_panitia tidak boleh mereview event miliknya
	// sendiri, sekalipun (secara hipotetis) berhasil mencapai endpoint ini.
	// Defense-in-depth di luar RBAC role-gate di router.
	if event.OrganizerID == reviewerID {
		httpx.WriteError(w, http.StatusForbidden, "you cannot review your own event")
		return
	}

	if err := eventapproval.ValidateTransition(event.Status, targetStatus); err != nil {
		httpx.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	now := time.Now()

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(event).Updates(map[string]interface{}{
			"status":     targetStatus,
			"updated_by": reviewerID,
		}).Error; err != nil {
			return err
		}

		log := models.EventApprovalLog{
			EventID:      id,
			EventVersion: event.EventVersion,
			Action:       logAction,
			ReviewedByID: &reviewerID,
			ReviewedAt:   &now,
			Notes:        req.Notes,
		}
		return tx.Create(&log).Error
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to process review action")
		return
	}

	event.Status = targetStatus
	httpx.WriteJSON(w, http.StatusOK, event)
}

// ApproveEvent menyetujui event yang sedang pending_approval.
// POST /api/events/{id}/approve (khusus super_admin/school_reviewer)
func (h *EventHandler) ApproveEvent(w http.ResponseWriter, r *http.Request) {
	h.reviewAction(w, r, models.EventStatusApproved, models.ApprovalActionApproved, false)
}

// RejectEvent menolak event secara final. Notes/alasan WAJIB diisi.
// POST /api/events/{id}/reject (khusus super_admin/school_reviewer)
func (h *EventHandler) RejectEvent(w http.ResponseWriter, r *http.Request) {
	h.reviewAction(w, r, models.EventStatusRejected, models.ApprovalActionRejected, true)
}

// RequestRevision meminta Admin Panitia memperbaiki event. Notes WAJIB diisi.
// POST /api/events/{id}/request-revision (khusus super_admin/school_reviewer)
func (h *EventHandler) RequestRevision(w http.ResponseWriter, r *http.Request) {
	h.reviewAction(w, r, models.EventStatusRevisionRequested, models.ApprovalActionRevisionRequested, true)
}

// PublishEvent mempublikasikan event yang sudah approved.
//
// PENTING: aksi ini SENGAJA TIDAK menulis baris ke event_approval_logs —
// kolom `action` di tabel itu punya CHECK constraint yang cuma mengizinkan
// submitted/approved/rejected/revision_requested (lihat models/
// event_approval_log.go). Audit trail publish cukup mengandalkan
// Event.PublishedAt & Event.Status.
//
// POST /api/events/{id}/publish (khusus super_admin)
func (h *EventHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	event, err := h.loadEvent(id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}

	if err := eventapproval.ValidateTransition(event.Status, models.EventStatusPublished); err != nil {
		httpx.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	now := time.Now()
	if err := h.DB.Model(event).Updates(map[string]interface{}{
		"status":       models.EventStatusPublished,
		"published_at": now,
		"updated_by":   userID,
	}).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to publish event")
		return
	}

	event.Status = models.EventStatusPublished
	event.PublishedAt = &now
	httpx.WriteJSON(w, http.StatusOK, event)
}

// UnpublishEvent menarik kembali event yang sudah published, balik ke draft.
//
// KEPUTUSAN PROJECT: status dikembalikan ke "draft" (bukan "approved").
// Event yang di-unpublish harus melalui ulang siklus submit -> approve
// sebelum bisa dipublish lagi. Sama seperti publish, TIDAK menulis ke
// event_approval_logs (lihat catatan di PublishEvent).
//
// POST /api/events/{id}/unpublish (khusus super_admin)
func (h *EventHandler) UnpublishEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	event, err := h.loadEvent(id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}

	if err := eventapproval.ValidateTransition(event.Status, models.EventStatusDraft); err != nil {
		httpx.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.DB.Model(event).Updates(map[string]interface{}{
		"status":       models.EventStatusDraft,
		"published_at": nil,
		"updated_by":   userID,
	}).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to unpublish event")
		return
	}

	event.Status = models.EventStatusDraft
	event.PublishedAt = nil
	httpx.WriteJSON(w, http.StatusOK, event)
}
