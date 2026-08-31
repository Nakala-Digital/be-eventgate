package models

// Participant merepresentasikan entitas tabel `participants` dalam basis data.
// Entitas ini menyimpan informasi detail mengenai peserta yang mendaftar pada suatu event.
// Skema ini dikonfigurasi untuk tidak memiliki kolom `created_at` dan `updated_at`.
// Validasi unik diterapkan pada kombinasi `email` dan `name` untuk mencegah duplikasi entri;
// jika terdapat pendaftaran dengan kombinasi yang sama, sistem akan menggunakan kembali ID yang sudah ada.
type Participant struct {
	ID              uint   `gorm:"primaryKey;column:participant_id" json:"id"`
	Name            string `gorm:"column:name;size:255;not null" json:"name"`
	Email           string `gorm:"column:email;size:100;not null" json:"email"`
	PhoneNumber     string `gorm:"column:phone_number;size:50" json:"phone_number,omitempty"`
	StudentID       string `gorm:"column:student_id;size:50" json:"student_id,omitempty"`
	ClassName       string `gorm:"column:class_name;size:50" json:"class_name,omitempty"`
	GuardianName    string `gorm:"column:guardian_name;size:255" json:"guardian_name,omitempty"`
	InstitutionUnit string `gorm:"column:institution_unit;size:255" json:"institution_unit,omitempty"`
}

func (Participant) TableName() string { return "participants" }
