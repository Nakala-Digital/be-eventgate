// Package testutil berisi helper YANG HANYA DIPAKAI OLEH TEST, tidak pernah
// diimpor oleh kode production (main.go / router.go). Karena itu dependency
// gorm.io/driver/sqlite di sini tidak ikut ke binary production, hanya ke
// binary test.
package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"be-eventgate/internal/auth"
	"be-eventgate/internal/database"
	"be-eventgate/internal/models"
)

var eventSeq uint64

// ResponseEnvelope merepresentasikan kontrak respons JSON seluruh endpoint.
type ResponseEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Errors  json.RawMessage `json:"errors"`
}

// DecodeEnvelope memvalidasi keberadaan semua field envelope dan mengembalikan
// isi respons untuk dipakai oleh pengujian endpoint.
func DecodeEnvelope(t *testing.T, body []byte) ResponseEnvelope {
	t.Helper()

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("failed to decode JSON envelope: %v; body: %s", err, body)
	}
	for _, field := range []string{"success", "message", "data", "errors"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("JSON envelope is missing %q; body: %s", field, body)
		}
	}

	var envelope ResponseEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("failed to decode response envelope: %v; body: %s", err, body)
	}
	return envelope
}

// DecodeData mengambil payload dari field data pada JSON envelope.
func DecodeData[T any](t *testing.T, body []byte) T {
	t.Helper()
	envelope := DecodeEnvelope(t, body)

	var data T
	if len(envelope.Data) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Data), []byte("null")) {
		return data
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("failed to decode response data: %v; data: %s", err, envelope.Data)
	}
	return data
}

// MustSetupDB membuat instance SQLite in-memory yang unik per test (supaya
// antar test tidak saling mengganggu), sudah dimigrasi, dan sudah di-seed
// role. Gagal langsung t.Fatal kalau ada error.
func MustSetupDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}
	if err := database.SeedRoles(db); err != nil {
		t.Fatalf("failed to seed roles: %v", err)
	}
	return db
}

// CreateTestUser membuat user dengan password sudah di-hash dan role dicari
// berdasarkan nama role (harus sudah ada, mis. hasil SeedRoles).
func CreateTestUser(db *gorm.DB, username, email, plainPassword, roleName string, isActive bool) (*models.User, error) {
	var role models.Role
	if err := db.Where("role_name = ?", roleName).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role %q not found, did you call MustSetupDB (which seeds roles)? %w", roleName, err)
	}

	hashed, err := auth.HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	user := models.User{
		RoleID:   role.ID,
		Username: username,
		Email:    email,
		Password: hashed,
		IsActive: isActive,
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}
	user.Role = role
	return &user, nil
}

func CreateTestEvent(db *gorm.DB, organizerID uint, status string) (*models.Event, error) {
	now := time.Now()
	seq := atomic.AddUint64(&eventSeq, 1)
	event := models.Event{
		OrganizerID: organizerID,
		Title:       fmt.Sprintf("Test Event %d-%d", now.UnixNano(), seq),
		Description: "Test Description",
		Banner:      "http://example.com/banner.jpg",
		Location:    "Test Location",
		Slug:        fmt.Sprintf("test-event-%d-%d", now.UnixNano(), seq),
		StartTime:   now.Add(24 * time.Hour),
		EndTime:     now.Add(48 * time.Hour),
		IsPaid:      false,
		Price:       0,
		Quota:       100,
		Status:      status,
		CreatedByID: organizerID,
	}
	if err := db.Create(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func CreateTestTicketType(db *gorm.DB, eventID uint) (*models.TicketType, error) {
	tt := models.TicketType{
		EventID:     eventID,
		Name:        "Regular",
		IsPaid:      false,
		Price:       0,
		MaxCapacity: 100,
		SoldCount:   0,
		IsActive:    true,
	}
	if err := db.Create(&tt).Error; err != nil {
		return nil, err
	}
	return &tt, nil
}

// CreateTestTicketTypeWithOptions sama seperti CreateTestTicketType tapi
// dengan kontrol penuh atas IsPaid/Price/MaxCapacity.
func CreateTestTicketTypeWithOptions(db *gorm.DB, eventID uint, isPaid bool, price float64, maxCapacity int) (*models.TicketType, error) {
	tt := models.TicketType{
		EventID:     eventID,
		Name:        "Regular",
		IsPaid:      isPaid,
		Price:       price,
		MaxCapacity: maxCapacity,
		SoldCount:   0,
		IsActive:    true,
	}
	if err := db.Create(&tt).Error; err != nil {
		return nil, err
	}
	return &tt, nil
}

// CreateTestQuestion membuat satu entitas pertanyaan form dinamis yang dikaitkan dengan kegiatan tertentu.
func CreateTestQuestion(db *gorm.DB, eventID uint, questionType, requirementType string) (*models.DynamicQuestion, error) {
	q := models.DynamicQuestion{
		EventID:         eventID,
		QuestionText:    "Test Question",
		QuestionType:    questionType,
		RequirementType: requirementType,
		DisplayOrder:    0,
		IsActive:        true,
	}
	if err := db.Create(&q).Error; err != nil {
		return nil, err
	}
	return &q, nil
}
