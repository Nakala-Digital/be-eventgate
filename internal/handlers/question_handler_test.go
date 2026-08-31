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

func newQuestionRequest(method, url string, userID uint, roleName string, body interface{}) *http.Request {
	return newEventRequest(method, url, userID, roleName, body)
}

// ---- Create ----

func TestCreateQuestion_Success_TextType(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q1", "panitia_q1@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{
		QuestionText: "Nama lengkap?",
		QuestionType: models.QuestionTypeText,
	}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[handlers.QuestionResponse](t, rr.Body.Bytes())
	if resp.RequirementType != models.RequirementTypeWajib {
		t.Errorf("expected default requirement_type 'wajib', got %s", resp.RequirementType)
	}
	if resp.EventID != event.ID {
		t.Errorf("expected event_id %d, got %d", event.ID, resp.EventID)
	}
}

func TestCreateQuestion_Success_SelectWithOptions(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q2", "panitia_q2@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{
		QuestionText: "Ukuran baju?",
		QuestionType: models.QuestionTypeSelect,
		Options: []handlers.OptionInput{
			{OptionLabel: "S", OptionValue: "S", DisplayOrder: 1},
			{OptionLabel: "M", OptionValue: "M", DisplayOrder: 2},
			{OptionLabel: "L", OptionValue: "L", DisplayOrder: 3},
		},
	}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[handlers.QuestionResponse](t, rr.Body.Bytes())
	if len(resp.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(resp.Options))
	}
}

func TestCreateQuestion_MissingQuestionText_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q3", "panitia_q3@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{QuestionType: models.QuestionTypeText}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when question_text is missing, got %d", rr.Code)
	}
}

func TestCreateQuestion_DropdownType_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q4", "panitia_q4@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{QuestionText: "Test?", QuestionType: "dropdown"}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid question_type 'dropdown', got %d", rr.Code)
	}
}

func TestCreateQuestion_FileUploadType_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q5", "panitia_q5@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{QuestionText: "Upload KTP", QuestionType: "file_upload"}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported file_upload type, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateQuestion_OptionsOnTextType_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q6", "panitia_q6@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{
		QuestionText: "Test?",
		QuestionType: models.QuestionTypeText,
		Options:      []handlers.OptionInput{{OptionLabel: "A", OptionValue: "a"}},
	}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when options are given for a 'text' question, got %d", rr.Code)
	}
}

func TestCreateQuestion_SelectWithoutOptions_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q7", "panitia_q7@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{QuestionText: "Pilih?", QuestionType: models.QuestionTypeSelect}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when a select question has zero options, got %d", rr.Code)
	}
}

func TestCreateQuestion_Kondisional_RequiresDependsOn(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q8", "panitia_q8@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{
		QuestionText:    "Kondisional tanpa depends_on",
		QuestionType:    models.QuestionTypeText,
		RequirementType: models.RequirementTypeKondisional,
	}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when requirement_type is 'kondisional' without depends_on fields, got %d", rr.Code)
	}
}

func TestCreateQuestion_Kondisional_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q9", "panitia_q9@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	parentQ, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeRadio, models.RequirementTypeWajib)

	h := handlers.NewQuestionHandler(db)
	parentID := parentQ.ID
	body := handlers.QuestionRequest{
		QuestionText:        "Kalau iya, sebutkan alergi apa",
		QuestionType:        models.QuestionTypeText,
		RequirementType:     models.RequirementTypeKondisional,
		DependsOnQuestionID: &parentID,
		DependsOnValue:      "ya",
	}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 for valid kondisional question, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateQuestion_DependsOnDifferentEvent_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q10", "panitia_q10@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	eventA, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	eventB, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	questionInEventB, _ := testutil.CreateTestQuestion(db, eventB.ID, models.QuestionTypeRadio, models.RequirementTypeWajib)

	h := handlers.NewQuestionHandler(db)
	dependsID := questionInEventB.ID
	body := handlers.QuestionRequest{
		QuestionText:        "Cross event depends_on",
		QuestionType:        models.QuestionTypeText,
		RequirementType:     models.RequirementTypeKondisional,
		DependsOnQuestionID: &dependsID,
		DependsOnValue:      "x",
	}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", eventA.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", eventA.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when depends_on_question_id refers to a question from a different event, got %d", rr.Code)
	}
}

func TestCreateQuestion_NotOwner_Forbidden(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q11", "panitia_q11@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	otherPanitia, _ := testutil.CreateTestUser(db, "panitia_q12", "panitia_q12@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{QuestionText: "Test?", QuestionType: models.QuestionTypeText}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), otherPanitia.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when a non-owner admin_panitia tries to create a question, got %d", rr.Code)
	}
}

func TestCreateQuestion_SuperAdmin_CanCreateForAnyEvent(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q13", "panitia_q13@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	superAdmin, _ := testutil.CreateTestUser(db, "superadmin_q1", "superadmin_q1@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{QuestionText: "Test?", QuestionType: models.QuestionTypeText}
	req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), superAdmin.ID, models.RoleSuperAdmin, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 when super_admin creates a question for any event, got %d", rr.Code)
	}
}

func TestCreateQuestion_DisplayOrder_AutoIncrements(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q14", "panitia_q14@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)

	for i := 0; i < 3; i++ {
		body := handlers.QuestionRequest{QuestionText: fmt.Sprintf("Q%d", i), QuestionType: models.QuestionTypeText}
		req := newQuestionRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, body)
		req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 creating question %d, got %d", i, rr.Code)
		}
		resp := testutil.DecodeData[handlers.QuestionResponse](t, rr.Body.Bytes())
		if resp.DisplayOrder != i {
			t.Errorf("expected display_order %d for question %d, got %d", i, i, resp.DisplayOrder)
		}
	}
}

// ---- Update ----

func TestUpdateQuestion_ReplacesOptions(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q15", "panitia_q15@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	question, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeSelect, models.RequirementTypeWajib)
	db.Create(&models.QuestionOption{QuestionID: question.ID, OptionLabel: "Old", OptionValue: "old", IsActive: true})

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{
		QuestionText: "Updated question",
		QuestionType: models.QuestionTypeSelect,
		Options: []handlers.OptionInput{
			{OptionLabel: "New1", OptionValue: "new1"},
			{OptionLabel: "New2", OptionValue: "new2"},
		},
	}
	req := newQuestionRequest(http.MethodPut, fmt.Sprintf("/api/events/%d/questions/%d", event.ID, question.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	req = withChiParam(req, "questionID", fmt.Sprintf("%d", question.ID))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[handlers.QuestionResponse](t, rr.Body.Bytes())
	if len(resp.Options) != 2 {
		t.Fatalf("expected exactly 2 options after replace, got %d", len(resp.Options))
	}

	var count int64
	db.Model(&models.QuestionOption{}).Where("question_id = ? AND option_value = ?", question.ID, "old").Count(&count)
	if count != 0 {
		t.Fatalf("expected old option to be removed after replace, but it still exists")
	}
}

func TestUpdateQuestion_WrongEvent_NotFound(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q16", "panitia_q16@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	eventA, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	eventB, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	questionInB, _ := testutil.CreateTestQuestion(db, eventB.ID, models.QuestionTypeText, models.RequirementTypeWajib)

	h := handlers.NewQuestionHandler(db)
	body := handlers.QuestionRequest{QuestionText: "Test", QuestionType: models.QuestionTypeText}
	// Menguji pembaruan entitas pertanyaan dari kegiatan lain melalui rute yang tidak sesuai
	req := newQuestionRequest(http.MethodPut, fmt.Sprintf("/api/events/%d/questions/%d", eventA.ID, questionInB.ID), organizer.ID, models.RoleAdminPanitia, body)
	req = withChiParam(req, "id", fmt.Sprintf("%d", eventA.ID))
	req = withChiParam(req, "questionID", fmt.Sprintf("%d", questionInB.ID))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when question belongs to a different event, got %d", rr.Code)
	}
}

// ---- Delete (soft delete) ----

func TestDeleteQuestion_SoftDeletesAndExcludedFromDefaultList(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q17", "panitia_q17@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	question, _ := testutil.CreateTestQuestion(db, event.ID, models.QuestionTypeText, models.RequirementTypeWajib)

	h := handlers.NewQuestionHandler(db)
	delReq := newQuestionRequest(http.MethodDelete, fmt.Sprintf("/api/events/%d/questions/%d", event.ID, question.ID), organizer.ID, models.RoleAdminPanitia, nil)
	delReq = withChiParam(delReq, "id", fmt.Sprintf("%d", event.ID))
	delReq = withChiParam(delReq, "questionID", fmt.Sprintf("%d", question.ID))
	delRR := httptest.NewRecorder()
	h.Delete(delRR, delReq)

	if delRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", delRR.Code, delRR.Body.String())
	}

	// Data tetap ada di basis data (penghapusan logis), status is_active menjadi false
	var stillExists models.DynamicQuestion
	if err := db.First(&stillExists, question.ID).Error; err != nil {
		t.Fatalf("expected question row to still exist after soft-delete, got error: %v", err)
	}
	if stillExists.IsActive {
		t.Fatal("expected is_active to be false after delete")
	}

	// Parameter baku tidak menampilkan pertanyaan yang telah dinonaktifkan
	listReq := newQuestionRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	listReq = withChiParam(listReq, "id", fmt.Sprintf("%d", event.ID))
	listRR := httptest.NewRecorder()
	h.ListByEvent(listRR, listReq)

	list := testutil.DecodeData[[]handlers.QuestionResponse](t, listRR.Body.Bytes())
	if len(list) != 0 {
		t.Fatalf("expected 0 questions in default list after soft-delete, got %d", len(list))
	}
}

// ---- List authorization ----

func TestListQuestions_StafLapangan_Forbidden(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q18", "panitia_q18@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	staf, _ := testutil.CreateTestUser(db, "staf_q1", "staf_q1@eventgate.test", "Password123!", models.RoleStafLapangan, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewQuestionHandler(db)
	req := newQuestionRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/questions", event.ID), staf.ID, models.RoleStafLapangan, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ListByEvent(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for staf_lapangan listing questions, got %d", rr.Code)
	}
}

func TestListQuestions_OrderedByDisplayOrder(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia_q19", "panitia_q19@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	db.Create(&models.DynamicQuestion{EventID: event.ID, QuestionText: "Third", QuestionType: models.QuestionTypeText, RequirementType: models.RequirementTypeWajib, DisplayOrder: 3, IsActive: true})
	db.Create(&models.DynamicQuestion{EventID: event.ID, QuestionText: "First", QuestionType: models.QuestionTypeText, RequirementType: models.RequirementTypeWajib, DisplayOrder: 1, IsActive: true})
	db.Create(&models.DynamicQuestion{EventID: event.ID, QuestionText: "Second", QuestionType: models.QuestionTypeText, RequirementType: models.RequirementTypeWajib, DisplayOrder: 2, IsActive: true})

	h := handlers.NewQuestionHandler(db)
	req := newQuestionRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/questions", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ListByEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	list := testutil.DecodeData[[]handlers.QuestionResponse](t, rr.Body.Bytes())
	if len(list) != 3 {
		t.Fatalf("expected 3 questions, got %d", len(list))
	}
	if list[0].QuestionText != "First" || list[1].QuestionText != "Second" || list[2].QuestionText != "Third" {
		t.Errorf("expected order [First, Second, Third], got [%s, %s, %s]", list[0].QuestionText, list[1].QuestionText, list[2].QuestionText)
	}
}
