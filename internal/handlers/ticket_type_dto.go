package handlers

import (
	"time"

	"be-eventgate/internal/models"
)

// TicketTypeRequest merepresentasikan data masukan untuk pembuatan dan pembaruan tipe tiket.
type TicketTypeRequest struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	Quota       *int     `json:"quota"`
	MaxCapacity *int     `json:"max_capacity"`
	Description string   `json:"description"`
	IsPaid      *bool    `json:"is_paid"`
	IsActive    *bool    `json:"is_active"`
}

// TicketTypeResponse merepresentasikan struktur data balasan tipe tiket untuk klien terautentikasi / manajemen.
type TicketTypeResponse struct {
	ID          uint      `json:"id"`
	EventID     uint      `json:"event_id"`
	Name        string    `json:"name"`
	IsPaid      bool      `json:"is_paid"`
	Price       float64   `json:"price"`
	Quota       int       `json:"quota"`
	MaxCapacity int       `json:"max_capacity"`
	SoldCount   int       `json:"sold_count"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toTicketTypeResponse(t models.TicketType) TicketTypeResponse {
	return TicketTypeResponse{
		ID:          t.ID,
		EventID:     t.EventID,
		Name:        t.Name,
		IsPaid:      t.IsPaid,
		Price:       t.Price,
		Quota:       t.MaxCapacity,
		MaxCapacity: t.MaxCapacity,
		SoldCount:   t.SoldCount,
		Description: t.Description,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
