package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"be-eventgate/internal/httpx"
	"be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
)

type EventHandler struct {
	DB *gorm.DB
}

func NewEventHandler(db *gorm.DB) *EventHandler {
	return &EventHandler{DB: db}
}

type EventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Banner      string    `json:"banner"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Location    string    `json:"location"`
	IsPaid      bool      `json:"is_paid"`
	Price       float64   `json:"price"`
	Quota       int       `json:"quota"`
	Status      string    `json:"status"`
}

// slugify mengubah judul menjadi slug unik URL-friendly.
func slugify(title string) string {
	slug := strings.ToLower(title)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "event"
	}
	return fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%100000)
}

// isValidStatus menguji validitas status event.
func isValidStatus(status string) bool {
	switch status {
	case models.EventStatusDraft,
		models.EventStatusPendingApproval,
		models.EventStatusApproved,
		models.EventStatusRevisionRequested,
		models.EventStatusPublished,
		models.EventStatusRejected,
		models.EventStatusCancelled,
		models.EventStatusCompleted:
		return true
	default:
		return false
	}
}

// Create memproses pembuatan event baru.
// Endpoint: POST /api/events
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validasi field wajib
	if strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Description) == "" ||
		strings.TrimSpace(req.Banner) == "" ||
		strings.TrimSpace(req.Location) == "" ||
		req.StartTime.IsZero() ||
		req.EndTime.IsZero() {
		httpx.WriteError(w, http.StatusBadRequest, "title, description, banner, location, start_time, and end_time are required")
		return
	}

	if req.Quota < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "quota must be greater than or equal to 0")
		return
	}

	if !req.EndTime.After(req.StartTime) {
		httpx.WriteError(w, http.StatusBadRequest, "end_time must be after start_time")
		return
	}

	// Validasi event berbayar
	if req.IsPaid {
		if req.Price <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "price is required and must be greater than 0 for paid events")
			return
		}
	} else {
		req.Price = 0
	}

	// Status awal event
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = models.EventStatusDraft
	} else if !isValidStatus(status) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event status")
		return
	}

	event := models.Event{
		OrganizerID: userID,
		CreatedByID: userID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Banner:      strings.TrimSpace(req.Banner),
		Location:    strings.TrimSpace(req.Location),
		Slug:        slugify(req.Title),
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsPaid:      req.IsPaid,
		Price:       req.Price,
		Quota:       req.Quota,
		Status:      status,
	}

	if err := h.DB.Create(&event).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, event)
}

// List menampilkan daftar event dengan dukungan filter pencarian dan status.
// Endpoint: GET /api/events
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	query := h.DB.Model(&models.Event{})

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var events []models.Event
	if err := query.Order("created_at DESC").Find(&events).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch events")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, events)
}

// GetByID menampilkan rincian detail dari satu event.
// Endpoint: GET /api/events/{id}
func (h *EventHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	eventID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var event models.Event
	if err := h.DB.First(&event, uint(eventID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch event")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, event)
}

// Update mengidentifikasi dan memperbarui data event yang tersimpan.
// Endpoint: PUT /api/events/{id}
func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleName, _ := middleware.GetRoleName(r.Context())

	idParam := chi.URLParam(r, "id")
	eventID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var event models.Event
	if err := h.DB.First(&event, uint(eventID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch event")
		return
	}

	// Otorisasi Hak Akses RBAC:
	// Super Admin bebas mengubah event manapun.
	// Admin Panitia hanya diizinkan mengubah event miliknya sendiri.
	if roleName != models.RoleSuperAdmin && event.CreatedByID != userID && event.OrganizerID != userID {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to update this event")
		return
	}

	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Description) == "" ||
		strings.TrimSpace(req.Banner) == "" ||
		strings.TrimSpace(req.Location) == "" ||
		req.StartTime.IsZero() ||
		req.EndTime.IsZero() {
		httpx.WriteError(w, http.StatusBadRequest, "title, description, banner, location, start_time, and end_time are required")
		return
	}

	if req.Quota < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "quota must be greater than or equal to 0")
		return
	}

	if !req.EndTime.After(req.StartTime) {
		httpx.WriteError(w, http.StatusBadRequest, "end_time must be after start_time")
		return
	}

	if req.IsPaid {
		if req.Price <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "price is required and must be greater than 0 for paid events")
			return
		}
	} else {
		req.Price = 0
	}

	status := strings.TrimSpace(req.Status)
	if status != "" && !isValidStatus(status) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event status")
		return
	}

	if strings.TrimSpace(req.Title) != event.Title {
		event.Slug = slugify(req.Title)
	}

	event.Title = strings.TrimSpace(req.Title)
	event.Description = strings.TrimSpace(req.Description)
	event.Banner = strings.TrimSpace(req.Banner)
	event.Location = strings.TrimSpace(req.Location)
	event.StartTime = req.StartTime
	event.EndTime = req.EndTime
	event.IsPaid = req.IsPaid
	event.Price = req.Price
	event.Quota = req.Quota
	if status != "" {
		event.Status = status
	}
	event.UpdatedByID = &userID

	if err := h.DB.Save(&event).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update event")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, event)
}

// Delete menghapus (soft-delete) data event.
// Endpoint: DELETE /api/events/{id}
func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleName, _ := middleware.GetRoleName(r.Context())

	idParam := chi.URLParam(r, "id")
	eventID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var event models.Event
	if err := h.DB.First(&event, uint(eventID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch event")
		return
	}

	// Otorisasi Hak Akses RBAC:
	if roleName != models.RoleSuperAdmin && event.CreatedByID != userID && event.OrganizerID != userID {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to delete this event")
		return
	}

	if err := h.DB.Delete(&event).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete event")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "event deleted successfully"})
}
