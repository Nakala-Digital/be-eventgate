package models

import (
	"time"

	"gorm.io/gorm"
)

// Event status constants
const (
	EventStatusDraft             = "draft"
	EventStatusPendingApproval   = "pending_approval"
	EventStatusApproved          = "approved"
	EventStatusRevisionRequested = "revision_requested"
	EventStatusPublished         = "published"
	EventStatusRejected          = "rejected"
	EventStatusCancelled         = "cancelled"
	EventStatusCompleted         = "completed"
)

// Event mendefinisikan skema entitas acara/event dalam basis data.
type Event struct {
	ID          uint           `gorm:"primaryKey;column:event_id;autoIncrement" json:"id"`
	OrganizerID uint           `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	Organizer   User           `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
	Title       string         `gorm:"column:title;size:255;not null" json:"title"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Location    string         `gorm:"column:location;size:255;not null" json:"location"`
	Banner      string         `gorm:"column:banner;size:255" json:"banner"`
	Slug        string         `gorm:"column:slug;size:255;uniqueIndex" json:"slug"`
	StartTime   time.Time      `gorm:"column:start_time;not null" json:"start_time"`
	EndTime     time.Time      `gorm:"column:end_time;not null" json:"end_time"`
	IsPaid      bool           `gorm:"column:is_paid;not null;default:false" json:"is_paid"`
	Price       float64        `gorm:"column:price;type:numeric(12,2);not null;default:0" json:"price"`
	Quota       int            `gorm:"column:quota;not null;default:0" json:"quota"`
	Status      string         `gorm:"column:status;size:50;not null;default:'draft'" json:"status"`
	CreatedByID uint           `gorm:"column:created_by;not null;index" json:"created_by"`
	CreatedBy   User           `gorm:"foreignKey:CreatedByID" json:"created_by_user,omitempty"`
	UpdatedByID *uint          `gorm:"column:updated_by" json:"updated_by,omitempty"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

func (Event) TableName() string { return "events" }
