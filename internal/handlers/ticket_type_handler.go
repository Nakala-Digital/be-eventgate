package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"be-eventgate/internal/httpx"
	"be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
)

type TicketTypeHandler struct {
	DB *gorm.DB
}

func NewTicketTypeHandler(db *gorm.DB) *TicketTypeHandler {
	return &TicketTypeHandler{DB: db}
}

// canManageEventTicketTypes memeriksa kewenangan pengelolaan tiket kegiatan.
// Peran super_admin memiliki akses penuh, sedangkan admin_panitia hanya
// diizinkan mengelola tiket dari kegiatan miliknya.
func canManageEventTicketTypes(roleName string, userID uint, event *models.Event) bool {
	if roleName == models.RoleSuperAdmin {
		return true
	}
	if roleName == models.RoleAdminPanitia && (event.OrganizerID == userID || event.CreatedByID == userID) {
		return true
	}
	return false
}

func (h *TicketTypeHandler) loadEvent(id uint) (*models.Event, error) {
	var event models.Event
	if err := h.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// ListByEvent menampilkan seluruh tipe tiket untuk sebuah event (termasuk yang tidak aktif jika untuk manajemen).
// GET /api/events/{id}/ticket-types
func (h *TicketTypeHandler) ListByEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.loadEvent(eventID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load event")
		return
	}

	roleName, _ := middleware.GetRoleName(r.Context())
	userID, _ := middleware.GetUserID(r.Context())
	if !canViewEvent(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to view ticket types for this event")
		return
	}

	var ticketTypes []models.TicketType
	if err := h.DB.Where("event_id = ?", eventID).Order("ticket_type_id asc").Find(&ticketTypes).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load ticket types")
		return
	}

	resp := make([]TicketTypeResponse, 0, len(ticketTypes))
	for _, tt := range ticketTypes {
		resp = append(resp, toTicketTypeResponse(tt))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// Create menambahkan tipe tiket baru ke dalam suatu event.
// POST /api/events/{id}/ticket-types
func (h *TicketTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.loadEvent(eventID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load event")
		return
	}

	roleName, _ := middleware.GetRoleName(r.Context())
	userID, _ := middleware.GetUserID(r.Context())
	if !canManageEventTicketTypes(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to add ticket types to this event")
		return
	}

	var req TicketTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	capacity := 0
	if req.MaxCapacity != nil {
		capacity = *req.MaxCapacity
	} else if req.Quota != nil {
		capacity = *req.Quota
	}
	if capacity < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "quota / max_capacity must be greater than or equal to 0")
		return
	}

	price := 0.0
	if req.Price != nil {
		price = *req.Price
	}
	if price < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "price cannot be negative")
		return
	}

	isPaid := false
	if req.IsPaid != nil {
		isPaid = *req.IsPaid
	} else if price > 0 {
		isPaid = true
	}

	if isPaid && price <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "price must be greater than 0 for paid ticket type")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	ticketType := models.TicketType{
		EventID:     eventID,
		Name:        name,
		IsPaid:      isPaid,
		Price:       price,
		MaxCapacity: capacity,
		SoldCount:   0,
		Description: strings.TrimSpace(req.Description),
		IsActive:    isActive,
	}

	if err := h.DB.Create(&ticketType).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create ticket type")
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, "ticket type created successfully", toTicketTypeResponse(ticketType))
}

// Update memperbarui konfigurasi tipe tiket yang sudah ada.
// PUT /api/events/{id}/ticket-types/{ticketTypeID}
func (h *TicketTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	ticketTypeID, err := parseURLUintParam(r, "ticketTypeID")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid ticket type id")
		return
	}

	event, err := h.loadEvent(eventID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load event")
		return
	}

	roleName, _ := middleware.GetRoleName(r.Context())
	userID, _ := middleware.GetUserID(r.Context())
	if !canManageEventTicketTypes(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to update ticket types for this event")
		return
	}

	var ticketType models.TicketType
	if err := h.DB.First(&ticketType, ticketTypeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "ticket type not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load ticket type")
		return
	}

	if ticketType.EventID != eventID {
		httpx.WriteError(w, http.StatusBadRequest, "ticket type does not belong to this event")
		return
	}

	var req TicketTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" {
		ticketType.Name = strings.TrimSpace(req.Name)
	}

	if req.MaxCapacity != nil {
		if *req.MaxCapacity < ticketType.SoldCount {
			httpx.WriteError(w, http.StatusBadRequest, "max_capacity cannot be less than current sold_count")
			return
		}
		ticketType.MaxCapacity = *req.MaxCapacity
	} else if req.Quota != nil {
		if *req.Quota < ticketType.SoldCount {
			httpx.WriteError(w, http.StatusBadRequest, "quota cannot be less than current sold_count")
			return
		}
		ticketType.MaxCapacity = *req.Quota
	}

	if req.Price != nil {
		if *req.Price < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "price cannot be negative")
			return
		}
		ticketType.Price = *req.Price
		if req.IsPaid == nil {
			ticketType.IsPaid = *req.Price > 0
		}
	}

	if req.IsPaid != nil {
		ticketType.IsPaid = *req.IsPaid
		if ticketType.IsPaid && ticketType.Price <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "price must be greater than 0 for paid ticket type")
			return
		}
	}

	if req.IsActive != nil {
		ticketType.IsActive = *req.IsActive
	}

	if req.Description != "" {
		ticketType.Description = strings.TrimSpace(req.Description)
	}

	if err := h.DB.Save(&ticketType).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update ticket type")
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "ticket type updated successfully", toTicketTypeResponse(ticketType))
}

// Delete menghapus tipe tiket dari event.
// DELETE /api/events/{id}/ticket-types/{ticketTypeID}
func (h *TicketTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	ticketTypeID, err := parseURLUintParam(r, "ticketTypeID")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid ticket type id")
		return
	}

	event, err := h.loadEvent(eventID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load event")
		return
	}

	roleName, _ := middleware.GetRoleName(r.Context())
	userID, _ := middleware.GetUserID(r.Context())
	if !canManageEventTicketTypes(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to delete ticket types for this event")
		return
	}

	var ticketType models.TicketType
	if err := h.DB.First(&ticketType, ticketTypeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "ticket type not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load ticket type")
		return
	}

	if ticketType.EventID != eventID {
		httpx.WriteError(w, http.StatusBadRequest, "ticket type does not belong to this event")
		return
	}

	if ticketType.SoldCount > 0 {
		httpx.WriteError(w, http.StatusConflict, "cannot delete ticket type with sold tickets")
		return
	}

	var regCount int64
	if err := h.DB.Model(&models.Registration{}).Where("ticket_type_id = ?", ticketTypeID).Count(&regCount).Error; err == nil && regCount > 0 {
		httpx.WriteError(w, http.StatusConflict, "cannot delete ticket type with existing registrations")
		return
	}

	if err := h.DB.Delete(&ticketType).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete ticket type")
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "ticket type deleted successfully", nil)
}
