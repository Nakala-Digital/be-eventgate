package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"be-eventgate/internal/httpx"
	"be-eventgate/internal/models"
)

type RegistrationHandler struct {
	DB *gorm.DB
}

func NewRegistrationHandler(db *gorm.DB) *RegistrationHandler {
	return &RegistrationHandler{DB: db}
}

func parsePublicEventIDParam(r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *RegistrationHandler) loadPublishedEvent(eventID uint) (*models.Event, error) {
	var event models.Event
	if err := h.DB.First(&event, eventID).Error; err != nil {
		return nil, err
	}
	if event.Status != models.EventStatusPublished {
		return nil, gorm.ErrRecordNotFound // sengaja disamarkan jadi "not found" ke publik
	}
	return &event, nil
}

// ---- Endpoint Publik Prasyarat ----
// Endpoint berikut diperlukan agar klien publik dapat memperoleh referensi
// data tiket dan form sebelum melakukan registrasi.

// ListPublicTicketTypes mengembalikan daftar tipe tiket yang berstatus aktif
// untuk suatu event yang telah dipublikasikan. Endpoint ini bersifat publik.
// GET /api/public/events/{id}/ticket-types
func (h *RegistrationHandler) ListPublicTicketTypes(w http.ResponseWriter, r *http.Request) {
	eventID, err := parsePublicEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	if _, err := h.loadPublishedEvent(eventID); err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found or not open for registration")
		return
	}

	var ticketTypes []models.TicketType
	if err := h.DB.Where("event_id = ? AND is_active = ?", eventID, true).Order("ticket_type_id asc").Find(&ticketTypes).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load ticket types")
		return
	}

	resp := make([]PublicTicketTypeResponse, 0, len(ticketTypes))
	for _, t := range ticketTypes {
		resp = append(resp, toPublicTicketTypeResponse(t))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// ListPublicQuestions mengembalikan struktur form (pertanyaan dinamis) yang berstatus aktif
// untuk suatu event yang telah dipublikasikan. Endpoint ini bersifat publik.
// GET /api/public/events/{id}/questions
func (h *RegistrationHandler) ListPublicQuestions(w http.ResponseWriter, r *http.Request) {
	eventID, err := parsePublicEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	if _, err := h.loadPublishedEvent(eventID); err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found or not open for registration")
		return
	}

	var questions []models.DynamicQuestion
	if err := h.DB.Where("event_id = ? AND is_active = ?", eventID, true).
		Preload("Options").
		Order("display_order asc, question_id asc").
		Find(&questions).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load questions")
		return
	}

	resp := make([]QuestionResponse, 0, len(questions))
	for _, q := range questions {
		resp = append(resp, toQuestionResponse(q))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// ---- Registrasi ----

func generateRegistrationCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // tanpa 0/O/1/I biar gak ambigu dibaca
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return "REG-" + string(b)
}

func (h *RegistrationHandler) uniqueRegistrationCode() (string, error) {
	for i := 0; i < 20; i++ {
		code := generateRegistrationCode()
		var count int64
		if err := h.DB.Model(&models.Registration{}).Where("registration_code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate a unique registration code after several attempts")
}

// isAnswerEmpty mengevaluasi apakah sebuah input jawaban dianggap kosong.
// Jawaban dianggap kosong apabila bernilai null, string kosong, atau array kosong.
func isAnswerEmpty(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || trimmed == `""` || trimmed == "[]" {
		return true
	}
	return false
}

// answerMatchesValue memverifikasi apakah jawaban tunggal atau elemen dalam array jawaban
// sesuai dengan nilai referensi tertentu. Fungsi ini utamanya digunakan untuk evaluasi
// logika pertanyaan kondisional (depends_on_value).
func answerMatchesValue(raw json.RawMessage, value string) bool {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s == value
	}
	// mungkin dikirim sebagai array (mis. checkbox) — cocok kalau value ada di dalamnya
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		for _, v := range arr {
			if v == value {
				return true
			}
		}
	}
	return false
}

// validateOptionAnswer memvalidasi input jawaban untuk tipe dropdown, radio, dan checkbox
// untuk memastikan bahwa nilai yang dikirimkan terdaftar dalam opsi yang valid.
func validateOptionAnswer(q models.DynamicQuestion, raw json.RawMessage) error {
	validValues := make(map[string]bool, len(q.Options))
	for _, o := range q.Options {
		validValues[o.OptionValue] = true
	}

	if q.QuestionType == models.QuestionTypeCheckbox {
		var arr []string
		if err := json.Unmarshal(raw, &arr); err != nil {
			return fmt.Errorf("answer for question %d must be an array of strings", q.ID)
		}
		for _, v := range arr {
			if !validValues[v] {
				return fmt.Errorf("answer %q for question %d is not a valid option", v, q.ID)
			}
		}
		return nil
	}

	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return fmt.Errorf("answer for question %d must be a string", q.ID)
	}
	if !validValues[s] {
		return fmt.Errorf("answer %q for question %d is not a valid option", s, q.ID)
	}
	return nil
}

// Register memproses pendaftaran peserta baru ke dalam suatu event.
// Endpoint ini bersifat publik dan memvalidasi ketersediaan tiket serta input dinamis.
// POST /api/public/events/{id}/register
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	eventID, err := parsePublicEventIDParam(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.loadPublishedEvent(eventID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "event not found or not open for registration")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Participant.Name = strings.TrimSpace(req.Participant.Name)
	req.Participant.Email = strings.ToLower(strings.TrimSpace(req.Participant.Email))
	if req.Participant.Name == "" || req.Participant.Email == "" {
		httpx.WriteError(w, http.StatusBadRequest, "participant.name and participant.email are required")
		return
	}
	if req.TicketTypeID == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "ticket_type_id is required")
		return
	}

	var ticketType models.TicketType
	if err := h.DB.First(&ticketType, req.TicketTypeID).Error; err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "ticket_type_id does not refer to an existing ticket type")
		return
	}
	if ticketType.EventID != event.ID {
		httpx.WriteError(w, http.StatusBadRequest, "ticket_type_id does not belong to this event")
		return
	}
	if !ticketType.IsActive {
		httpx.WriteError(w, http.StatusConflict, "this ticket type is no longer available")
		return
	}
	if ticketType.SoldCount >= ticketType.MaxCapacity {
		httpx.WriteError(w, http.StatusConflict, "this ticket type is sold out")
		return
	}

	// Membatasi registrasi ganda: satu alamat email hanya diperbolehkan memiliki
	// satu registrasi aktif per event (registrasi dengan status dibatalkan tidak dihitung).
	var duplicateCount int64
	h.DB.Model(&models.Registration{}).
		Joins("JOIN participants ON participants.participant_id = registrations.participant_id").
		Where("participants.email = ? AND registrations.event_id = ? AND registrations.status != ?",
			req.Participant.Email, event.ID, models.RegistrationStatusCancelled).
		Count(&duplicateCount)
	if duplicateCount > 0 {
		httpx.WriteError(w, http.StatusConflict, "this email has already registered for this event")
		return
	}

	// Memuat seluruh pertanyaan dinamis yang aktif pada event ini untuk keperluan validasi.
	var questions []models.DynamicQuestion
	if err := h.DB.Where("event_id = ? AND is_active = ?", eventID, true).Preload("Options").Find(&questions).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load event questions")
		return
	}
	questionByID := make(map[uint]models.DynamicQuestion, len(questions))
	for _, q := range questions {
		questionByID[q.ID] = q
	}

	answerByQuestionID := make(map[uint]json.RawMessage, len(req.Answers))
	for _, a := range req.Answers {
		q, ok := questionByID[a.QuestionID]
		if !ok {
			httpx.WriteError(w, http.StatusBadRequest, fmt.Sprintf("question_id %d does not belong to this event or is inactive", a.QuestionID))
			return
		}
		if models.OptionSupportingQuestionTypes[q.QuestionType] && !isAnswerEmpty(a.Answer) {
			if err := validateOptionAnswer(q, a.Answer); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
		}
		answerByQuestionID[a.QuestionID] = a.Answer
	}

	// Memvalidasi kelengkapan jawaban wajib:
	// - Tipe 'wajib' harus selalu diisi.
	// - Tipe 'kondisional' menjadi wajib jika jawaban dari pertanyaan acuan sesuai.
	for _, q := range questions {
		required := q.RequirementType == models.RequirementTypeWajib
		if q.RequirementType == models.RequirementTypeKondisional && q.DependsOnQuestionID != nil {
			parentAnswer, parentAnswered := answerByQuestionID[*q.DependsOnQuestionID]
			if parentAnswered && answerMatchesValue(parentAnswer, q.DependsOnValue) {
				required = true
			}
		}
		if required && isAnswerEmpty(answerByQuestionID[q.ID]) {
			httpx.WriteError(w, http.StatusBadRequest, fmt.Sprintf("question_id %d is required but was not answered", q.ID))
			return
		}
	}

	registrationCode, err := h.uniqueRegistrationCode()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to generate registration code")
		return
	}

	status := models.RegistrationStatusWaitingPayment
	var confirmedAt *time.Time
	if !ticketType.IsPaid {
		status = models.RegistrationStatusConfirmed
		now := time.Now()
		confirmedAt = &now
	}

	var registration models.Registration
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Mencari entitas peserta yang sudah ada berdasarkan kombinasi unik nama dan email.
		// Jika tidak ditemukan, sistem akan membuat entitas peserta baru.
		var participant models.Participant
		err := tx.Where("email = ? AND name = ?", req.Participant.Email, req.Participant.Name).First(&participant).Error
		if err == gorm.ErrRecordNotFound {
			participant = models.Participant{
				Name: req.Participant.Name, Email: req.Participant.Email,
				PhoneNumber: req.Participant.PhoneNumber, StudentID: req.Participant.StudentID,
				ClassName: req.Participant.ClassName, GuardianName: req.Participant.GuardianName,
				InstitutionUnit: req.Participant.InstitutionUnit,
			}
			if err := tx.Create(&participant).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// Mengurangi kapasitas kuota tiket secara atomik untuk mencegah overselling
		// akibat race condition saat pendaftaran dilakukan secara serentak.
		// Pembaruan hanya berhasil jika kapasitas yang terjual masih di bawah kapasitas maksimum.
		result := tx.Model(&models.TicketType{}).
			Where("ticket_type_id = ? AND sold_count < max_capacity", ticketType.ID).
			UpdateColumn("sold_count", gorm.Expr("sold_count + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("ticket type is sold out")
		}

		registration = models.Registration{
			RegistrationCode: registrationCode,
			ParticipantID:    participant.ID,
			EventID:          event.ID,
			TicketTypeID:     ticketType.ID,
			RegistrationDate: time.Now(),
			Status:           status,
			ConfirmedAt:      confirmedAt,
		}
		if err := tx.Create(&registration).Error; err != nil {
			return err
		}
		registration.Participant = participant

		for questionID, answer := range answerByQuestionID {
			if isAnswerEmpty(answer) {
				continue // tidak menyimpan baris untuk pertanyaan yang tidak dijawab
			}
			fr := models.FormResponse{
				RegistrationID: registration.ID,
				QuestionID:     questionID,
				AnswerJSON:     answer,
			}
			if err := tx.Create(&fr).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err.Error() == "ticket type is sold out" {
			httpx.WriteError(w, http.StatusConflict, "this ticket type is sold out")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create registration")
		return
	}

	var savedAnswers []models.FormResponse
	h.DB.Where("registration_id = ?", registration.ID).Find(&savedAnswers)

	httpx.WriteJSON(w, http.StatusCreated, toRegistrationResponse(registration, savedAnswers))
}

// GetByCode mengembalikan detail registrasi berdasarkan registration_code.
// Kode ini berfungsi sebagai kunci akses unik bagi peserta publik.
// Endpoint ini bersifat publik.
// GET /api/public/registrations/{code}
func (h *RegistrationHandler) GetByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		httpx.WriteError(w, http.StatusBadRequest, "registration code is required")
		return
	}

	var registration models.Registration
	if err := h.DB.Preload("Participant").Where("registration_code = ?", code).First(&registration).Error; err != nil {
		httpx.WriteError(w, http.StatusNotFound, "registration not found")
		return
	}

	var answers []models.FormResponse
	h.DB.Where("registration_id = ?", registration.ID).Find(&answers)

	httpx.WriteJSON(w, http.StatusOK, toRegistrationResponse(registration, answers))
}
