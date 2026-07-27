package models

import (
	"encoding/json"
	"time"
)

// FormResponse merepresentasikan entitas tabel `form_responses` dalam basis data.
// Entitas ini menyimpan data jawaban dinamis dari form registrasi peserta.
// Entitas ini bersifat append-only (hanya simpan baru) dan tidak mendukung pembaruan,
// sehingga hanya memiliki kolom `created_at`.
// Terdapat batasan unik pada kombinasi `registration_id` dan `question_id`
// untuk memastikan setiap pertanyaan hanya dijawab maksimal satu kali per registrasi.
type FormResponse struct {
	ID             uint            `gorm:"primaryKey;column:response_id" json:"id"`
	RegistrationID uint            `gorm:"column:registration_id;not null;index" json:"registration_id"`
	QuestionID     uint            `gorm:"column:question_id;not null;index" json:"question_id"`
	AnswerJSON     json.RawMessage `gorm:"column:answer_json;type:jsonb;not null" json:"answer"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (FormResponse) TableName() string { return "form_responses" }
