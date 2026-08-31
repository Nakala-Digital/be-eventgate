package models

import "time"

// Konstanta berikut merepresentasikan status registrasi peserta dan harus
// selaras dengan CHECK constraint pada skema basis data.
const (
	RegistrationStatusPending        = "pending"
	RegistrationStatusPendingPayment = "pending_payment"
	RegistrationStatusConfirmed      = "confirmed"
	RegistrationStatusCancelled      = "cancelled"
)

// Registration merepresentasikan tabel `registrations`.
type Registration struct {
	ID               uint        `gorm:"primaryKey;column:registration_id" json:"id"`
	RegistrationCode string      `gorm:"column:registration_code;size:100;not null;uniqueIndex" json:"registration_code"`
	ParticipantID    uint        `gorm:"column:participant_id;not null;index" json:"participant_id"`
	Participant      Participant `json:"participant,omitempty"`
	EventID          uint        `gorm:"column:event_id;not null;index" json:"event_id"`
	TicketTypeID     uint        `gorm:"column:ticket_type_id;not null;index" json:"ticket_type_id"`
	TicketType       TicketType  `json:"ticket_type,omitempty"`
	RegistrationDate time.Time   `gorm:"column:registration_date;not null" json:"registration_date"`
	Status           string      `gorm:"column:status;size:50;not null" json:"status"`
	ConfirmedAt      *time.Time  `gorm:"column:confirmed_at" json:"confirmed_at,omitempty"`
	CancelledAt      *time.Time  `gorm:"column:cancelled_at" json:"cancelled_at,omitempty"`
	CancelledReason  string      `gorm:"column:cancelled_reason;type:text" json:"cancelled_reason,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

func (Registration) TableName() string { return "registrations" }
