package handlers

import (
	"encoding/json"
	"time"

	"be-eventgate/internal/models"
)

// ParticipantInput merepresentasikan struktur data masukan yang dikirim oleh klien
// pada saat proses pendaftaran (registrasi) peserta baru.
type ParticipantInput struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number,omitempty"`
	StudentID       string `json:"student_id,omitempty"`
	ClassName       string `json:"class_name,omitempty"`
	GuardianName    string `json:"guardian_name,omitempty"`
	InstitutionUnit string `json:"institution_unit,omitempty"`
}

// AnswerInput merepresentasikan struktur data masukan untuk jawaban spesifik
// terhadap pertanyaan dinamis dalam form pendaftaran.
type AnswerInput struct {
	QuestionID uint            `json:"question_id"`
	Answer     json.RawMessage `json:"answer"`
}

// RegisterRequest merepresentasikan struktur _payload_ utama yang diharapkan dari klien
// untuk endpoint pendaftaran peserta publik.
type RegisterRequest struct {
	Participant  ParticipantInput `json:"participant"`
	TicketTypeID uint             `json:"ticket_type_id"`
	Answers      []AnswerInput    `json:"answers"`
}

type ParticipantResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number,omitempty"`
	StudentID       string `json:"student_id,omitempty"`
	ClassName       string `json:"class_name,omitempty"`
	GuardianName    string `json:"guardian_name,omitempty"`
	InstitutionUnit string `json:"institution_unit,omitempty"`
}

func toParticipantResponse(p models.Participant) ParticipantResponse {
	return ParticipantResponse{
		ID: p.ID, Name: p.Name, Email: p.Email, PhoneNumber: p.PhoneNumber,
		StudentID: p.StudentID, ClassName: p.ClassName,
		GuardianName: p.GuardianName, InstitutionUnit: p.InstitutionUnit,
	}
}

type AnswerResponse struct {
	QuestionID uint            `json:"question_id"`
	Answer     json.RawMessage `json:"answer"`
}

// RegistrationResponse merepresentasikan struktur data balasan (response) lengkap
// mengenai detail registrasi peserta.
type RegistrationResponse struct {
	ID               uint                `json:"id"`
	RegistrationCode string              `json:"registration_code"`
	EventID          uint                `json:"event_id"`
	TicketTypeID     uint                `json:"ticket_type_id"`
	Status           string              `json:"status"`
	RegistrationDate time.Time           `json:"registration_date"`
	ConfirmedAt      *time.Time          `json:"confirmed_at,omitempty"`
	CancelledAt      *time.Time          `json:"cancelled_at,omitempty"`
	CancelledReason  string              `json:"cancelled_reason,omitempty"`
	Participant      ParticipantResponse `json:"participant"`
	Answers          []AnswerResponse    `json:"answers,omitempty"`
}

func toRegistrationResponse(r models.Registration, answers []models.FormResponse) RegistrationResponse {
	answerResp := make([]AnswerResponse, 0, len(answers))
	for _, a := range answers {
		answerResp = append(answerResp, AnswerResponse{QuestionID: a.QuestionID, Answer: a.AnswerJSON})
	}
	return RegistrationResponse{
		ID:               r.ID,
		RegistrationCode: r.RegistrationCode,
		EventID:          r.EventID,
		TicketTypeID:     r.TicketTypeID,
		Status:           r.Status,
		RegistrationDate: r.RegistrationDate,
		ConfirmedAt:      r.ConfirmedAt,
		CancelledAt:      r.CancelledAt,
		CancelledReason:  r.CancelledReason,
		Participant:      toParticipantResponse(r.Participant),
		Answers:          answerResp,
	}
}

// PublicTicketTypeResponse merepresentasikan struktur data balasan tiket untuk klien publik.
// Struktur ini secara sengaja mengecualikan kolom-kolom internal (seperti sold_count murni)
// dan hanya memberikan informasi kuota tersisa (RemainingQuota).
type PublicTicketTypeResponse struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	IsPaid         bool    `json:"is_paid"`
	Price          float64 `json:"price"`
	RemainingQuota int     `json:"remaining_quota"`
	Description    string  `json:"description,omitempty"`
}

func toPublicTicketTypeResponse(t models.TicketType) PublicTicketTypeResponse {
	remaining := t.MaxCapacity - t.SoldCount
	if remaining < 0 {
		remaining = 0
	}
	return PublicTicketTypeResponse{
		ID: t.ID, Name: t.Name, IsPaid: t.IsPaid, Price: t.Price,
		RemainingQuota: remaining, Description: t.Description,
	}
}
