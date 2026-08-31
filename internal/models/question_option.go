package models

// QuestionOption merepresentasikan entitas opsi pertanyaan pada basis data.
// Entitas ini tidak menggunakan kolom pencatatan waktu standar.
type QuestionOption struct {
	ID           uint   `gorm:"primaryKey;column:option_id" json:"id"`
	QuestionID   uint   `gorm:"column:question_id;not null;index" json:"question_id"`
	OptionLabel  string `gorm:"column:option_label;size:255;not null" json:"option_label"`
	OptionValue  string `gorm:"column:option_value;size:255;not null" json:"option_value"`
	DisplayOrder int    `gorm:"column:display_order;not null" json:"display_order"`
	IsActive     bool   `gorm:"column:is_active;not null" json:"is_active"`
}

func (QuestionOption) TableName() string { return "question_options" }
