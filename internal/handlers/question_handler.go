package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"be-eventgate/internal/httpx"
	"be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
)

type QuestionHandler struct {
	DB *gorm.DB
}

func NewQuestionHandler(db *gorm.DB) *QuestionHandler {
	return &QuestionHandler{DB: db}
}

// canManageEventQuestions memeriksa kewenangan pengelolaan pertanyaan suatu kegiatan.
// Peran super_admin memiliki akses penuh, sedangkan admin_panitia hanya
// diizinkan mengelola pertanyaan dari kegiatan yang dimilikinya.
func canManageEventQuestions(roleName string, userID uint, event *models.Event) bool {
	if roleName == models.RoleSuperAdmin {
		return true
	}
	if roleName == models.RoleAdminPanitia && event.OrganizerID == userID {
		return true
	}
	return false
}

func isValidQuestionType(t string) bool {
	for _, v := range models.ValidQuestionTypes {
		if v == t {
			return true
		}
	}
	return false
}

func isValidRequirementType(t string) bool {
	for _, v := range models.ValidRequirementTypes {
		if v == t {
			return true
		}
	}
	return false
}

func parseURLUintParam(r *http.Request, key string) (uint, error) {
	idStr := chi.URLParam(r, key)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *QuestionHandler) loadEvent(id uint) (*models.Event, error) {
	var event models.Event
	if err := h.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// validateQuestionRequest menjalankan proses validasi aturan bisnis untuk
// pembuatan atau pembaruan pertanyaan, di luar dari pemeriksaan rekaman yang sudah ada.
// Fungsi ini dimanfaatkan oleh fungsi Create maupun Update.
func (h *QuestionHandler) validateQuestionRequest(eventID uint, req QuestionRequest, excludeQuestionID *uint) (string, int, string) {
	// Mengembalikan (pesanError, statusHTTP, "") jika terdapat kesalahan validasi; ("", 0, "") jika valid
	if req.QuestionText == "" {
		return "question_text is required", http.StatusBadRequest, ""
	}
	if req.QuestionType == "" || !isValidQuestionType(req.QuestionType) {
		return "question_type is required and must be one of: text, textarea, number, date, select, radio, checkbox", http.StatusBadRequest, ""
	}

	requirementType := req.RequirementType
	if requirementType == "" {
		requirementType = models.RequirementTypeWajib
	}
	if !isValidRequirementType(requirementType) {
		return "requirement_type must be one of: wajib, opsional, kondisional", http.StatusBadRequest, ""
	}

	if requirementType == models.RequirementTypeKondisional {
		if req.DependsOnQuestionID == nil || req.DependsOnValue == "" {
			return "depends_on_question_id and depends_on_value are required when requirement_type is 'kondisional'", http.StatusBadRequest, ""
		}
		var dependsOn models.DynamicQuestion
		if err := h.DB.First(&dependsOn, *req.DependsOnQuestionID).Error; err != nil {
			return "depends_on_question_id does not refer to an existing question", http.StatusBadRequest, ""
		}
		if dependsOn.EventID != eventID {
			return "depends_on_question_id must refer to a question belonging to the same event", http.StatusBadRequest, ""
		}
		if excludeQuestionID != nil && dependsOn.ID == *excludeQuestionID {
			return "a question cannot depend on itself", http.StatusBadRequest, ""
		}
	}

	if len(req.Options) > 0 && !models.OptionSupportingQuestionTypes[req.QuestionType] {
		return "options can only be set for question_type: select, radio, checkbox", http.StatusBadRequest, ""
	}
	if models.OptionSupportingQuestionTypes[req.QuestionType] && len(req.Options) == 0 {
		return "at least one option is required for question_type: select, radio, checkbox", http.StatusBadRequest, ""
	}
	for _, o := range req.Options {
		if o.OptionLabel == "" || o.OptionValue == "" {
			return "each option requires option_label and option_value", http.StatusBadRequest, ""
		}
	}

	return "", 0, ""
}

// ListByEvent mengambil seluruh daftar pertanyaan form dinamis yang terkait
// dengan suatu kegiatan tertentu, diurutkan berdasarkan field display_order.
//
// GET /api/events/{eventID}/questions?include_inactive=true
func (h *QuestionHandler) ListByEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	event, err := h.loadEvent(eventID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}

	roleName, _ := middleware.GetRoleName(r.Context())
	userID, _ := middleware.GetUserID(r.Context())
	if !canViewEvent(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to view this event's questions")
		return
	}

	query := h.DB.Where("event_id = ?", eventID).Preload("Options")
	if r.URL.Query().Get("include_inactive") != "true" {
		query = query.Where("is_active = ?", true)
	}

	var questions []models.DynamicQuestion
	if err := query.Order("display_order asc, question_id asc").Find(&questions).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load questions")
		return
	}

	resp := make([]QuestionResponse, 0, len(questions))
	for _, q := range questions {
		resp = append(resp, toQuestionResponse(q))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// Create membuat entitas pertanyaan form dinamis baru yang terhubung pada suatu kegiatan.
// POST /api/events/{eventID}/questions
func (h *QuestionHandler) Create(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleName, _ := middleware.GetRoleName(r.Context())

	event, err := h.loadEvent(eventID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}
	if !canManageEventQuestions(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to manage questions for this event")
		return
	}

	var req QuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg, status, _ := h.validateQuestionRequest(eventID, req, nil); msg != "" {
		httpx.WriteError(w, status, msg)
		return
	}

	requirementType := req.RequirementType
	if requirementType == "" {
		requirementType = models.RequirementTypeWajib
	}

	displayOrder := 0
	if req.DisplayOrder != nil {
		displayOrder = *req.DisplayOrder
	} else {
		var maxOrder *int
		h.DB.Model(&models.DynamicQuestion{}).Where("event_id = ?", eventID).Select("MAX(display_order)").Scan(&maxOrder)
		if maxOrder != nil {
			displayOrder = *maxOrder + 1
		}
	}

	question := models.DynamicQuestion{
		EventID:             eventID,
		QuestionText:        req.QuestionText,
		QuestionType:        req.QuestionType,
		RequirementType:     requirementType,
		DependsOnQuestionID: req.DependsOnQuestionID,
		DependsOnValue:      req.DependsOnValue,
		ValidationRule:      req.ValidationRule,
		Placeholder:         req.Placeholder,
		DisplayOrder:        displayOrder,
		IsActive:            true,
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&question).Error; err != nil {
			return err
		}
		for _, o := range req.Options {
			opt := models.QuestionOption{
				QuestionID:   question.ID,
				OptionLabel:  o.OptionLabel,
				OptionValue:  o.OptionValue,
				DisplayOrder: o.DisplayOrder,
				IsActive:     true,
			}
			if err := tx.Create(&opt).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create question")
		return
	}

	h.DB.Preload("Options").First(&question, question.ID)
	httpx.WriteJSON(w, http.StatusCreated, toQuestionResponse(question))
}

// Update memodifikasi data pertanyaan yang sudah ada.
// Apabila field opsi disertakan dalam request (meskipun sebagai array kosong),
// seluruh data opsi lama yang terkait akan ditimpa dengan kumpulan data yang baru.
// PUT /api/events/{eventID}/questions/{id}
func (h *QuestionHandler) Update(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	questionID, err := parseURLUintParam(r, "questionID")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid question id")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleName, _ := middleware.GetRoleName(r.Context())

	event, err := h.loadEvent(eventID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}
	if !canManageEventQuestions(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to manage questions for this event")
		return
	}

	var question models.DynamicQuestion
	if err := h.DB.Where("event_id = ?", eventID).First(&question, questionID).Error; err != nil {
		httpx.WriteError(w, http.StatusNotFound, "question not found for this event")
		return
	}

	var req QuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	excludeID := question.ID
	if msg, status, _ := h.validateQuestionRequest(eventID, req, &excludeID); msg != "" {
		httpx.WriteError(w, status, msg)
		return
	}

	requirementType := req.RequirementType
	if requirementType == "" {
		requirementType = models.RequirementTypeWajib
	}
	displayOrder := question.DisplayOrder
	if req.DisplayOrder != nil {
		displayOrder = *req.DisplayOrder
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&question).Updates(map[string]interface{}{
			"question_text":          req.QuestionText,
			"question_type":          req.QuestionType,
			"requirement_type":       requirementType,
			"depends_on_question_id": req.DependsOnQuestionID,
			"depends_on_value":       req.DependsOnValue,
			"validation_rule":        req.ValidationRule,
			"placeholder":            req.Placeholder,
			"display_order":          displayOrder,
		}).Error; err != nil {
			return err
		}

		if req.Options != nil {
			if err := tx.Where("question_id = ?", question.ID).Delete(&models.QuestionOption{}).Error; err != nil {
				return err
			}
			for _, o := range req.Options {
				opt := models.QuestionOption{
					QuestionID:   question.ID,
					OptionLabel:  o.OptionLabel,
					OptionValue:  o.OptionValue,
					DisplayOrder: o.DisplayOrder,
					IsActive:     true,
				}
				if err := tx.Create(&opt).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update question")
		return
	}

	h.DB.Preload("Options").First(&question, question.ID)
	httpx.WriteJSON(w, http.StatusOK, toQuestionResponse(question))
}

// Delete melakukan penghapusan data secara logis (menandai is_active menjadi false),
// bukan menghapus baris data secara fisik. Hal ini dilakukan karena tabel
// dynamic_questions tidak memiliki atribut pencatatan waktu penghapusan standar,
// serta untuk menjaga integritas referensial data jika terdapat relasi form_responses
// pada pertanyaan tersebut.
// DELETE /api/events/{eventID}/questions/{id}
func (h *QuestionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseURLUintParam(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	questionID, err := parseURLUintParam(r, "questionID")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid question id")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roleName, _ := middleware.GetRoleName(r.Context())

	event, err := h.loadEvent(eventID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found")
		return
	}
	if !canManageEventQuestions(roleName, userID, event) {
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to manage questions for this event")
		return
	}

	var question models.DynamicQuestion
	if err := h.DB.Where("event_id = ?", eventID).First(&question, questionID).Error; err != nil {
		httpx.WriteError(w, http.StatusNotFound, "question not found for this event")
		return
	}

	if err := h.DB.Model(&question).Update("is_active", false).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete question")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "question deleted successfully"})
}
