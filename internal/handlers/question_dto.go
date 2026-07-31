package handlers

import (
	"time"

	"be-eventgate/internal/models"
)

// OptionInput merepresentasikan data masukan opsi pertanyaan yang digunakan
// pada body request untuk pembuatan dan pembaruan pertanyaan.
type OptionInput struct {
	ID           *uint  `json:"id,omitempty"` // Opsional saat pembaruan, namun implementasi saat ini selalu menimpa seluruh opsi
	OptionLabel  string `json:"option_label"`
	OptionValue  string `json:"option_value"`
	DisplayOrder int    `json:"display_order"`
}

// QuestionRequest merepresentasikan data masukan untuk pembuatan dan pembaruan pertanyaan.
type QuestionRequest struct {
	QuestionText        string        `json:"question_text"`
	QuestionType        string        `json:"question_type"`
	RequirementType     string        `json:"requirement_type"` // Opsional, default: "wajib"
	DependsOnQuestionID *uint         `json:"depends_on_question_id,omitempty"`
	DependsOnValue      string        `json:"depends_on_value,omitempty"`
	ValidationRule      string        `json:"validation_rule,omitempty"`
	Placeholder         string        `json:"placeholder,omitempty"`
	DisplayOrder        *int          `json:"display_order,omitempty"` // Opsional, akan ditambahkan otomatis di urutan terakhir jika kosong
	Options             []OptionInput `json:"options,omitempty"`
}

type OptionResponse struct {
	ID           uint   `json:"id"`
	OptionLabel  string `json:"option_label"`
	OptionValue  string `json:"option_value"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
}

type QuestionResponse struct {
	ID                  uint             `json:"id"`
	EventID             uint             `json:"event_id"`
	QuestionText        string           `json:"question_text"`
	QuestionType        string           `json:"question_type"`
	RequirementType     string           `json:"requirement_type"`
	DependsOnQuestionID *uint            `json:"depends_on_question_id,omitempty"`
	DependsOnValue      string           `json:"depends_on_value,omitempty"`
	ValidationRule      string           `json:"validation_rule,omitempty"`
	Placeholder         string           `json:"placeholder,omitempty"`
	DisplayOrder        int              `json:"display_order"`
	IsActive            bool             `json:"is_active"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	Options             []OptionResponse `json:"options,omitempty"`
}

func toOptionResponse(o models.QuestionOption) OptionResponse {
	return OptionResponse{
		ID:           o.ID,
		OptionLabel:  o.OptionLabel,
		OptionValue:  o.OptionValue,
		DisplayOrder: o.DisplayOrder,
		IsActive:     o.IsActive,
	}
}

func toQuestionResponse(q models.DynamicQuestion) QuestionResponse {
	opts := make([]OptionResponse, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, toOptionResponse(o))
	}
	return QuestionResponse{
		ID:                  q.ID,
		EventID:             q.EventID,
		QuestionText:        q.QuestionText,
		QuestionType:        q.QuestionType,
		RequirementType:     q.RequirementType,
		DependsOnQuestionID: q.DependsOnQuestionID,
		DependsOnValue:      q.DependsOnValue,
		ValidationRule:      q.ValidationRule,
		Placeholder:         q.Placeholder,
		DisplayOrder:        q.DisplayOrder,
		IsActive:            q.IsActive,
		CreatedAt:           q.CreatedAt,
		UpdatedAt:           q.UpdatedAt,
		Options:             opts,
	}
}
