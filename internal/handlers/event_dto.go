package handlers

import (
	"time"

	"be-eventgate/internal/models"
)

// ApprovalLogResponse adalah representasi EventApprovalLog untuk client.
// TIDAK ADA field CreatedAt — tabel event_approval_logs memang tidak punya
// kolom itu (lihat catatan di models/event_approval_log.go). Urutan log
// diwakili oleh field ID (log_id, auto-increment berurutan).
type ApprovalLogResponse struct {
	ID            uint       `json:"id"`
	EventID       uint       `json:"event_id"`
	EventVersion  int        `json:"event_version"`
	Action        string     `json:"action"`
	SubmittedByID *uint      `json:"submitted_by,omitempty"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	ReviewedByID  *uint      `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	Notes         string     `json:"notes,omitempty"`
}

func toApprovalLogResponse(l models.EventApprovalLog) ApprovalLogResponse {
	return ApprovalLogResponse{
		ID:            l.ID,
		EventID:       l.EventID,
		EventVersion:  l.EventVersion,
		Action:        l.Action,
		SubmittedByID: l.SubmittedByID,
		SubmittedAt:   l.SubmittedAt,
		ReviewedByID:  l.ReviewedByID,
		ReviewedAt:    l.ReviewedAt,
		Notes:         l.Notes,
	}
}
