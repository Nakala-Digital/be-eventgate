package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	appdb "be-eventgate/internal/database"
	"be-eventgate/internal/handlers"
	"be-eventgate/internal/models"
	"be-eventgate/internal/testutil"
)

// TestRegister_ConcurrentRequests_DoesNotOversell menggunakan PostgreSQL
// karena SQLite in-memory tidak merepresentasikan locking dan isolation level
// yang dipakai pada deployment aplikasi. Jalankan pada database test khusus
// melalui EVENTGATE_TEST_POSTGRES_DSN.
func TestRegister_ConcurrentRequests_DoesNotOversell(t *testing.T) {
	dsn := os.Getenv("EVENTGATE_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set EVENTGATE_TEST_POSTGRES_DSN to run the PostgreSQL concurrency test")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open PostgreSQL test database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get PostgreSQL sql.DB: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(50)

	if err := appdb.Migrate(db); err != nil {
		t.Fatalf("failed to migrate PostgreSQL test database: %v", err)
	}
	if err := appdb.SeedRoles(db); err != nil {
		t.Fatalf("failed to seed PostgreSQL test database: %v", err)
	}

	suffix := time.Now().UnixNano()
	organizer, err := testutil.CreateTestUser(
		db,
		fmt.Sprintf("race_organizer_%d", suffix),
		fmt.Sprintf("race_organizer_%d@example.com", suffix),
		"Password123!",
		models.RoleAdminPanitia,
		true,
	)
	if err != nil {
		t.Fatalf("failed to create organizer: %v", err)
	}
	event, err := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	ticketType, err := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 5)
	if err != nil {
		t.Fatalf("failed to create ticket type: %v", err)
	}

	h := handlers.NewRegistrationHandler(db)
	const requestCount = 20
	results := make(chan struct {
		status int
		body   []byte
	}, requestCount)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(requestCount)

	for i := 0; i < requestCount; i++ {
		go func(i int) {
			defer wg.Done()
			<-start

			body := validRegisterBody(ticketType.ID, fmt.Sprintf("race_participant_%d_%d@example.com", suffix, i))
			req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
			req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
			rr := httptest.NewRecorder()
			h.Register(rr, req)
			results <- struct {
				status int
				body   []byte
			}{status: rr.Code, body: append([]byte(nil), rr.Body.Bytes()...)}
		}(i)
	}

	close(start)
	wg.Wait()
	close(results)

	successes := 0
	conflicts := 0
	for result := range results {
		testutil.DecodeEnvelope(t, result.body)
		switch result.status {
		case http.StatusCreated:
			successes++
		case http.StatusConflict:
			conflicts++
		default:
			t.Errorf("expected 201 or 409 under contention, got %d; body: %s", result.status, result.body)
		}
	}

	if successes != 5 {
		t.Errorf("expected exactly 5 successful registrations, got %d", successes)
	}
	if conflicts != requestCount-successes {
		t.Errorf("expected %d sold-out conflicts, got %d", requestCount-successes, conflicts)
	}

	var persistedTicketType models.TicketType
	if err := db.First(&persistedTicketType, ticketType.ID).Error; err != nil {
		t.Fatalf("failed to reload ticket type: %v", err)
	}
	if persistedTicketType.SoldCount != 5 {
		t.Errorf("expected sold_count to remain capped at 5, got %d", persistedTicketType.SoldCount)
	}

	var registrationCount int64
	if err := db.Model(&models.Registration{}).Where("ticket_type_id = ?", ticketType.ID).Count(&registrationCount).Error; err != nil {
		t.Fatalf("failed to count registrations: %v", err)
	}
	if registrationCount != 5 {
		t.Errorf("expected 5 persisted registrations, got %d", registrationCount)
	}
}
