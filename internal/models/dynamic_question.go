package models

import "time"

// Tipe field pertanyaan (question_type).
// Nilai ini harus selaras dengan batasan validasi pada basis data.
const (
	QuestionTypeText     = "text"
	QuestionTypeTextarea = "textarea"
	QuestionTypeNumber   = "number"
	QuestionTypeDate     = "date"
	QuestionTypeSelect   = "select"
	QuestionTypeRadio    = "radio"
	QuestionTypeCheckbox = "checkbox"
)

// ValidQuestionTypes memuat daftar seluruh tipe pertanyaan yang diizinkan.
var ValidQuestionTypes = []string{
	QuestionTypeText, QuestionTypeTextarea, QuestionTypeNumber, QuestionTypeDate,
	QuestionTypeSelect, QuestionTypeRadio, QuestionTypeCheckbox,
}

// OptionSupportingQuestionTypes mendaftar tipe pertanyaan yang memerlukan
// opsi tambahan secara fungsional, seperti select, radio, dan checkbox.
var OptionSupportingQuestionTypes = map[string]bool{
	QuestionTypeSelect:   true,
	QuestionTypeRadio:    true,
	QuestionTypeCheckbox: true,
}

// Konstanta tipe kewajiban pertanyaan (requirement_type).
// Tipe kondisional digunakan apabila pertanyaan ini hanya ditampilkan
// berdasarkan nilai jawaban dari pertanyaan lainnya.
const (
	RequirementTypeWajib       = "wajib"
	RequirementTypeOpsional    = "opsional"
	RequirementTypeKondisional = "kondisional"
)

var ValidRequirementTypes = []string{
	RequirementTypeWajib, RequirementTypeOpsional, RequirementTypeKondisional,
}

// DynamicQuestion merepresentasikan entitas pertanyaan dinamis pada basis data.
type DynamicQuestion struct {
	ID              uint   `gorm:"primaryKey;column:question_id" json:"id"`
	EventID         uint   `gorm:"column:event_id;not null;index" json:"event_id"`
	QuestionText    string `gorm:"column:question_text;type:text;not null" json:"question_text"`
	QuestionType    string `gorm:"column:question_type;size:50;not null" json:"question_type"`
	RequirementType string `gorm:"column:requirement_type;size:50;not null" json:"requirement_type"`

	// Referensi untuk pertanyaan kondisional. Pertanyaan ini akan ditampilkan
	// jika pertanyaan acuan dijawab sesuai dengan nilai pada DependsOnValue.
	DependsOnQuestionID *uint  `gorm:"column:depends_on_question_id" json:"depends_on_question_id,omitempty"`
	DependsOnValue      string `gorm:"column:depends_on_value;size:255" json:"depends_on_value,omitempty"`

	ValidationRule string `gorm:"column:validation_rule;size:255" json:"validation_rule,omitempty"`
	Placeholder    string `gorm:"column:placeholder;size:255" json:"placeholder,omitempty"`
	DisplayOrder   int    `gorm:"column:display_order;not null" json:"display_order"`
	IsActive       bool   `gorm:"column:is_active;not null" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Options []QuestionOption `gorm:"foreignKey:QuestionID" json:"options,omitempty"`
}

func (DynamicQuestion) TableName() string { return "dynamic_questions" }
