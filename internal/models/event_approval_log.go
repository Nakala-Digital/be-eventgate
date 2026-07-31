package models

import "time"

// Definisi konstanta untuk nilai aksi pada Event Approval Log.
// Nilai ini dibatasi oleh skema database, sehingga aksi seperti "published"
// atau "unpublished" tidak dicatat di dalam tabel log ini.
const (
	ApprovalActionSubmitted         = "submitted"
	ApprovalActionApproved          = "approved"
	ApprovalActionRejected          = "rejected"
	ApprovalActionRevisionRequested = "revision_requested"
)

// EventApprovalLog merepresentasikan entitas log persetujuan event di dalam database.
// Pengurutan log secara kronologis bergantung pada log_id.
type EventApprovalLog struct {
	ID           uint   `gorm:"primaryKey;column:log_id" json:"id"`
	EventID      uint   `gorm:"column:event_id;not null;index" json:"event_id"`
	EventVersion int    `gorm:"column:event_version;not null" json:"event_version"`
	Action       string `gorm:"column:action;size:50;not null" json:"action"`

	SubmittedByID *uint      `gorm:"column:submitted_by" json:"submitted_by,omitempty"`
	SubmittedAt   *time.Time `gorm:"column:submitted_at" json:"submitted_at,omitempty"`

	ReviewedByID *uint      `gorm:"column:reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`

	Notes string `gorm:"column:notes;type:text" json:"notes,omitempty"`
}

func (EventApprovalLog) TableName() string { return "event_approval_logs" }
