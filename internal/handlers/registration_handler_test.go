package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-eventgate/internal/handlers"
	"be-eventgate/internal/models"
	"be-eventgate/internal/testutil"
)

func newPublicRequest(method, url string, body interface{}) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader([]byte("{}"))
	}
	return httptest.NewRequest(method, url, reader)
}

func validRegisterBody(ticketTypeID uint, email string) handlers.RegisterRequest {
	return handlers.RegisterRequest{
		Participant: handlers.ParticipantInput{
			Name:  "Budi Santoso",
			Email: email,
		},
		TicketTypeID: ticketTypeID,
	}
}

// ---- ListPublicTicketTypes ----

func TestListPublicTicketTypes_PublishedEvent_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r1", "panitia_r1@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)
	req := newPublicRequest(http.MethodGet, fmt.Sprintf("/api/public/events/%d/ticket-types", event.ID), nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ListPublicTicketTypes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}
	var resp []handlers.PublicTicketTypeResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 ticket type, got %d", len(resp))
	}
	if resp[0].RemainingQuota != 100 {
		t.Errorf("expected remaining_quota 100, got %d", resp[0].RemainingQuota)
	}
}

func TestListPublicTicketTypes_UnpublishedEvent_NotFound(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r2", "panitia_r2@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)
	req := newPublicRequest(http.MethodGet, fmt.Sprintf("/api/public/events/%d/ticket-types", event.ID), nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ListPublicTicketTypes(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a draft (unpublished) event, got %d", rr.Code)
	}
}

// ---- ListPublicQuestions ----

func TestListPublicQuestions_OnlyActiveShown(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r3", "panitia_r3@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeText, models.RequirementTypeWajib)
	inactive, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeText, models.RequirementTypeWajib)
	db.Model(inactive).Update("is_active", false)

	h := handlers.NewRegistrationHandler(db)
	req := newPublicRequest(http.MethodGet, fmt.Sprintf("/api/public/events/%d/questions", event.ID), nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ListPublicQuestions(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp []handlers.QuestionResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 active question (inactive one excluded), got %d", len(resp))
	}
}

// ---- Register ----

func TestRegister_FreeTicket_ConfirmsImmediately(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r4", "panitia_r4@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)
	body := validRegisterBody(ticketType.ID, "budi1@example.com")
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", rr.Code, rr.Body.String())
	}
	var resp handlers.RegistrationResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Status != models.RegistrationStatusConfirmed {
		t.Errorf("expected status confirmed for a free ticket, got %s", resp.Status)
	}
	if resp.RegistrationCode == "" {
		t.Error("expected a non-empty registration_code")
	}
}

func TestRegister_PaidTicket_WaitingPayment(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r5", "panitia_r5@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, true, 50000, 100)

	h := handlers.NewRegistrationHandler(db)
	body := validRegisterBody(ticketType.ID, "budi2@example.com")
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", rr.Code, rr.Body.String())
	}
	var resp handlers.RegistrationResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Status != models.RegistrationStatusWaitingPayment {
		t.Errorf("expected status waiting_payment for a paid ticket, got %s", resp.Status)
	}
}

func TestRegister_UnpublishedEvent_NotFound(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r6", "panitia_r6@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)
	body := validRegisterBody(ticketType.ID, "budi3@example.com")
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when registering for an unpublished event, got %d", rr.Code)
	}
}

func TestRegister_TicketTypeFromDifferentEvent_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r7", "panitia_r7@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	eventA, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	eventB, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketTypeB, _ := testutil.CreateTestTicketTypeWithOptions(db, eventB.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)
	body := validRegisterBody(ticketTypeB.ID, "budi4@example.com")
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", eventA.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", eventA.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when ticket_type_id belongs to a different event, got %d", rr.Code)
	}
}

func TestRegister_SoldOut_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r8", "panitia_r8@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 1) // kapasitas cuma 1

	h := handlers.NewRegistrationHandler(db)

	firstBody := validRegisterBody(ticketType.ID, "first@example.com")
	firstReq := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), firstBody)
	firstReq = withChiParam(firstReq, "id", fmt.Sprintf("%d", event.ID))
	firstRR := httptest.NewRecorder()
	h.Register(firstRR, firstReq)
	if firstRR.Code != http.StatusCreated {
		t.Fatalf("first registration setup failed: expected 201, got %d, body: %s", firstRR.Code, firstRR.Body.String())
	}

	secondBody := validRegisterBody(ticketType.ID, "second@example.com")
	secondReq := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), secondBody)
	secondReq = withChiParam(secondReq, "id", fmt.Sprintf("%d", event.ID))
	secondRR := httptest.NewRecorder()
	h.Register(secondRR, secondReq)

	if secondRR.Code != http.StatusConflict {
		t.Fatalf("expected 409 when ticket type is sold out, got %d, body: %s", secondRR.Code, secondRR.Body.String())
	}
}

func TestRegister_DuplicateEmailSameEvent_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r9", "panitia_r9@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)

	body := validRegisterBody(ticketType.ID, "dup@example.com")
	req1 := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req1 = withChiParam(req1, "id", fmt.Sprintf("%d", event.ID))
	rr1 := httptest.NewRecorder()
	h.Register(rr1, req1)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("first registration setup failed: expected 201, got %d", rr1.Code)
	}

	req2 := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req2 = withChiParam(req2, "id", fmt.Sprintf("%d", event.ID))
	rr2 := httptest.NewRecorder()
	h.Register(rr2, req2)

	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for a duplicate email registering to the same event, got %d", rr2.Code)
	}
}

func TestRegister_MissingRequiredAnswer_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r10", "panitia_r10@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)
	testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeText, models.RequirementTypeWajib)

	h := handlers.NewRegistrationHandler(db)
	body := validRegisterBody(ticketType.ID, "budi5@example.com") // tidak ada answers sama sekali
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when a required question is not answered, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestRegister_WithAnswers_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r11", "panitia_r11@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)
	question, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeText, models.RequirementTypeWajib)

	h := handlers.NewRegistrationHandler(db)
	body := handlers.RegisterRequest{
		Participant:  handlers.ParticipantInput{Name: "Budi Answers", Email: "budi6@example.com"},
		TicketTypeID: ticketType.ID,
		Answers: []handlers.AnswerInput{
			{QuestionID: question.ID, Answer: json.RawMessage(`"Jawaban saya"`)},
		},
	}
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var count int64
	db.Model(&models.FormResponse{}).Where("question_id = ?", question.ID).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 form_response row to be saved, got %d", count)
	}
}

func TestRegister_InvalidDropdownAnswer_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r12", "panitia_r12@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)
	question, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeDropdown, models.RequirementTypeOpsional)
	db.Create(&models.QuestionOption{QuestionID: question.ID, OptionLabel: "S", OptionValue: "S", IsActive: true})
	db.Create(&models.QuestionOption{QuestionID: question.ID, OptionLabel: "M", OptionValue: "M", IsActive: true})

	h := handlers.NewRegistrationHandler(db)
	body := handlers.RegisterRequest{
		Participant:  handlers.ParticipantInput{Name: "Budi Invalid", Email: "budi7@example.com"},
		TicketTypeID: ticketType.ID,
		Answers: []handlers.AnswerInput{
			{QuestionID: question.ID, Answer: json.RawMessage(`"XL"`)}, // bukan option yang valid
		},
	}
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an answer that isn't a valid dropdown option, got %d", rr.Code)
	}
}

func TestRegister_KondisionalTriggered_RequiresAnswer(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r13", "panitia_r13@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	parentQ, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeRadio, models.RequirementTypeWajib)
	db.Create(&models.QuestionOption{QuestionID: parentQ.ID, OptionLabel: "Ya", OptionValue: "ya", IsActive: true})
	db.Create(&models.QuestionOption{QuestionID: parentQ.ID, OptionLabel: "Tidak", OptionValue: "tidak", IsActive: true})

	childQ := models.DynamicQuestion{
		EventID: event.ID, QuestionText: "Sebutkan alergi", QuestionType: models.QuestionTypeText,
		RequirementType: models.RequirementTypeKondisional, DependsOnQuestionID: &parentQ.ID, DependsOnValue: "ya",
		IsActive: true,
	}
	db.Create(&childQ)

	h := handlers.NewRegistrationHandler(db)
	body := handlers.RegisterRequest{
		Participant:  handlers.ParticipantInput{Name: "Budi Kondisional", Email: "budi8@example.com"},
		TicketTypeID: ticketType.ID,
		Answers: []handlers.AnswerInput{
			{QuestionID: parentQ.ID, Answer: json.RawMessage(`"ya"`)},
			// childQ TIDAK dijawab, padahal harusnya wajib karena parent == "ya"
		},
	}
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when a triggered kondisional question is not answered, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestRegister_KondisionalNotTriggered_NotRequired(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r14", "panitia_r14@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	parentQ, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeRadio, models.RequirementTypeWajib)
	db.Create(&models.QuestionOption{QuestionID: parentQ.ID, OptionLabel: "Ya", OptionValue: "ya", IsActive: true})
	db.Create(&models.QuestionOption{QuestionID: parentQ.ID, OptionLabel: "Tidak", OptionValue: "tidak", IsActive: true})

	childQ := models.DynamicQuestion{
		EventID: event.ID, QuestionText: "Sebutkan alergi", QuestionType: models.QuestionTypeText,
		RequirementType: models.RequirementTypeKondisional, DependsOnQuestionID: &parentQ.ID, DependsOnValue: "ya",
		IsActive: true,
	}
	db.Create(&childQ)

	h := handlers.NewRegistrationHandler(db)
	body := handlers.RegisterRequest{
		Participant:  handlers.ParticipantInput{Name: "Budi Kondisional2", Email: "budi9@example.com"},
		TicketTypeID: ticketType.ID,
		Answers: []handlers.AnswerInput{
			{QuestionID: parentQ.ID, Answer: json.RawMessage(`"tidak"`)},
			// childQ tidak dijawab -> OK karena parent-nya "tidak", bukan "ya"
		},
	}
	req := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 when a kondisional question is not triggered, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestRegister_SameParticipantDifferentEvent_ReusesParticipantRow(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r15", "panitia_r15@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	eventA, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	eventB, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketA, _ := testutil.CreateTestTicketTypeWithOptions(db, eventA.ID, false, 0, 100)
	ticketB, _ := testutil.CreateTestTicketTypeWithOptions(db, eventB.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)

	bodyA := validRegisterBody(ticketA.ID, "reuse@example.com")
	reqA := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", eventA.ID), bodyA)
	reqA = withChiParam(reqA, "id", fmt.Sprintf("%d", eventA.ID))
	rrA := httptest.NewRecorder()
	h.Register(rrA, reqA)
	if rrA.Code != http.StatusCreated {
		t.Fatalf("expected 201 for eventA registration, got %d", rrA.Code)
	}

	bodyB := validRegisterBody(ticketB.ID, "reuse@example.com")
	reqB := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", eventB.ID), bodyB)
	reqB = withChiParam(reqB, "id", fmt.Sprintf("%d", eventB.ID))
	rrB := httptest.NewRecorder()
	h.Register(rrB, reqB)
	if rrB.Code != http.StatusCreated {
		t.Fatalf("expected 201 for eventB registration (different event, same email should be allowed), got %d, body: %s", rrB.Code, rrB.Body.String())
	}

	var participantCount int64
	db.Model(&models.Participant{}).Where("email = ?", "reuse@example.com").Count(&participantCount)
	if participantCount != 1 {
		t.Fatalf("expected exactly 1 participant row to be reused across events, got %d", participantCount)
	}
}

// ---- GetByCode ----

func TestGetByCode_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_r16", "panitia_r16@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)
	ticketType, _ := testutil.CreateTestTicketTypeWithOptions(db, event.ID, false, 0, 100)

	h := handlers.NewRegistrationHandler(db)
	body := validRegisterBody(ticketType.ID, "getcode@example.com")
	regReq := newPublicRequest(http.MethodPost, fmt.Sprintf("/api/public/events/%d/register", event.ID), body)
	regReq = withChiParam(regReq, "id", fmt.Sprintf("%d", event.ID))
	regRR := httptest.NewRecorder()
	h.Register(regRR, regReq)
	var created handlers.RegistrationResponse
	_ = json.Unmarshal(regRR.Body.Bytes(), &created)

	getReq := newPublicRequest(http.MethodGet, fmt.Sprintf("/api/public/registrations/%s", created.RegistrationCode), nil)
	getReq = withChiParam(getReq, "code", created.RegistrationCode)
	getRR := httptest.NewRecorder()
	h.GetByCode(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", getRR.Code, getRR.Body.String())
	}
	var resp handlers.RegistrationResponse
	_ = json.Unmarshal(getRR.Body.Bytes(), &resp)
	if resp.RegistrationCode != created.RegistrationCode {
		t.Errorf("expected registration_code %s, got %s", created.RegistrationCode, resp.RegistrationCode)
	}
}

func TestGetByCode_NotFound(t *testing.T) {
	db := testutil.MustSetupDB(t)
	h := handlers.NewRegistrationHandler(db)

	req := newPublicRequest(http.MethodGet, "/api/public/registrations/REG-NOTEXIST", nil)
	req = withChiParam(req, "code", "REG-NOTEXIST")
	rr := httptest.NewRecorder()

	h.GetByCode(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a nonexistent registration code, got %d", rr.Code)
	}
}
