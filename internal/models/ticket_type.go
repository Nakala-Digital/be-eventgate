package models

import "time"

// TicketType merepresentasikan entitas jenis tiket untuk sebuah event di database.
type TicketType struct {
	ID          uint      `gorm:"primaryKey;column:ticket_type_id" json:"id"`
	EventID     uint      `gorm:"column:event_id;not null;index" json:"event_id"`
	Name        string    `gorm:"column:name;size:255;not null" json:"name"`
	IsPaid      bool      `gorm:"column:is_paid;not null" json:"is_paid"`
	Price       float64   `gorm:"column:price;type:numeric(12,2);not null" json:"price"`
	MaxCapacity int       `gorm:"column:max_capacity;not null" json:"max_capacity"`
	SoldCount   int       `gorm:"column:sold_count;not null" json:"sold_count"`
	Description string    `gorm:"column:description;type:text" json:"description,omitempty"`
	IsActive    bool      `gorm:"column:is_active;not null" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (TicketType) TableName() string { return "ticket_types" }
