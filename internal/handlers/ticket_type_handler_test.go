package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-eventgate/internal/handlers"
	"be-eventgate/internal/models"
	"be-eventgate/internal/testutil"
)

func newTicketTypeRequest(method, url string, userID uint, roleName string, body interface{}) *http.Request {
	return newEventRequest(method, url, userID, roleName, body)
}

// ---- Create Ticket Type ----

func TestCreateTicketType_Success_Free(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt1", "panitia_tt1@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewTicketTypeHandler(db)
	quota := 150
	price := 0.0
	body := handlers.TicketTypeRequest{
		Name:        "Early Bird Free",
		Quota:       &quota,
		Price:       &price,
		Description: "Tiket gratis untuk 150 pendaftar pertama",
	}
	req := newTicketTypeRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/ticket-types", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[handlers.TicketTypeResponse](t, rr.Body.Bytes())
	if resp.Name != "Early Bird Free" {
		t.Errorf("expected name 'Early Bird Free', got '%s'", resp.Name)
	}
	if resp.MaxCapacity != 150 || resp.Quota != 150 {
		t.Errorf("expected quota 150, got max_capacity=%d quota=%d", resp.MaxCapacity, resp.Quota)
	}
	if resp.IsPaid {
		t.Errorf("expected is_paid false for 0 price, got true")
	}
}

func TestCreateTicketType_Success_Paid(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt2", "panitia_tt2@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewTicketTypeHandler(db)
	quota := 50
	price := 75000.0
	body := handlers.TicketTypeRequest{
		Name:  "VIP",
		Quota: &quota,
		Price: &price,
	}
	req := newTicketTypeRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/ticket-types", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[handlers.TicketTypeResponse](t, rr.Body.Bytes())
	if !resp.IsPaid {
		t.Errorf("expected is_paid true for price > 0, got false")
	}
	if resp.Price != 75000 {
		t.Errorf("expected price 75000, got %f", resp.Price)
	}
}

func TestCreateTicketType_Validation_EmptyName(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt3", "panitia_tt3@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewTicketTypeHandler(db)
	body := handlers.TicketTypeRequest{
		Name: "",
	}
	req := newTicketTypeRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/ticket-types", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for empty name, got %d", rr.Code)
	}
}

func TestCreateTicketType_Validation_PaidWithoutPrice(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt4", "panitia_tt4@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewTicketTypeHandler(db)
	isPaid := true
	price := 0.0
	body := handlers.TicketTypeRequest{
		Name:   "Paid Ticket Invalid",
		IsPaid: &isPaid,
		Price:  &price,
	}
	req := newTicketTypeRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/ticket-types", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request when is_paid=true with price 0, got %d", rr.Code)
	}
}

func TestCreateTicketType_Forbidden_OtherPanitia(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "owner_tt", "owner_tt@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	otherPanitia, _ := testutil.CreateTestUser(db, "other_tt", "other_tt@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewTicketTypeHandler(db)
	body := handlers.TicketTypeRequest{
		Name: "Hacked Ticket",
	}
	req := newTicketTypeRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/ticket-types", event.ID), otherPanitia.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-owner panitia, got %d", rr.Code)
	}
}

// ---- List Ticket Types ----

func TestListTicketTypes_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt5", "panitia_tt5@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	testutil.CreateTestTicketType(db, event.ID)

	h := handlers.NewTicketTypeHandler(db)
	req := newTicketTypeRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/ticket-types", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ListByEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[[]handlers.TicketTypeResponse](t, rr.Body.Bytes())
	if len(resp) != 1 {
		t.Fatalf("expected 1 ticket type, got %d", len(resp))
	}
}

// ---- Update Ticket Type ----

func TestUpdateTicketType_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt6", "panitia_tt6@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	tt, _ := testutil.CreateTestTicketType(db, event.ID)

	h := handlers.NewTicketTypeHandler(db)
	newQuota := 200
	newPrice := 25000.0
	body := handlers.TicketTypeRequest{
		Name:  "Presale 2",
		Quota: &newQuota,
		Price: &newPrice,
	}
	req := newTicketTypeRequest(http.MethodPut, fmt.Sprintf("/api/events/%d/ticket-types/%d", event.ID, tt.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	req = withChiParam(req, "ticketTypeID", fmt.Sprintf("%d", tt.ID))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[handlers.TicketTypeResponse](t, rr.Body.Bytes())
	if resp.Name != "Presale 2" || resp.Price != 25000 || resp.Quota != 200 {
		t.Errorf("unexpected updated ticket type: %+v", resp)
	}
}

// ---- Delete Ticket Type ----

func TestDeleteTicketType_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt7", "panitia_tt7@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	tt, _ := testutil.CreateTestTicketType(db, event.ID)

	h := handlers.NewTicketTypeHandler(db)
	req := newTicketTypeRequest(http.MethodDelete, fmt.Sprintf("/api/events/%d/ticket-types/%d", event.ID, tt.ID), organizer.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	req = withChiParam(req, "ticketTypeID", fmt.Sprintf("%d", tt.ID))
	rr := httptest.NewRecorder()

	h.Delete(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var count int64
	db.Model(&models.TicketType{}).Where("ticket_type_id = ?", tt.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected ticket type to be deleted, found %d records", count)
	}
}

func TestDeleteTicketType_WithSoldTickets_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_tt8", "panitia_tt8@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	tt, _ := testutil.CreateTestTicketType(db, event.ID)

	// Set sold_count > 0
	db.Model(&models.TicketType{}).Where("ticket_type_id = ?", tt.ID).Update("sold_count", 5)

	h := handlers.NewTicketTypeHandler(db)
	req := newTicketTypeRequest(http.MethodDelete, fmt.Sprintf("/api/events/%d/ticket-types/%d", event.ID, tt.ID), organizer.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	req = withChiParam(req, "ticketTypeID", fmt.Sprintf("%d", tt.ID))
	rr := httptest.NewRecorder()

	h.Delete(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when deleting ticket type with sold tickets, got %d", rr.Code)
	}
}
