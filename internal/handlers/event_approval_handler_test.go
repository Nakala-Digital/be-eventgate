package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"be-eventgate/internal/handlers"
	"be-eventgate/internal/models"
	"be-eventgate/internal/testutil"
)

func newEventRequest(method, url string, userID uint, roleName string, body interface{}) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader([]byte("{}"))
	}
	req := httptest.NewRequest(method, url, reader)
	return req.WithContext(contextWithUser(req.Context(), userID, "user", roleName))
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.RouteContext(req.Context())
	if rctx == nil {
		rctx = chi.NewRouteContext()
	}
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

// ---- Submit for Approval ----

func TestSubmitForApproval_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia1", "panitia1@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, err := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	if err != nil {
		t.Fatalf("failed to create test event: %v", err)
	}
	if _, err := testutil.CreateTestTicketType(db, event.ID); err != nil {
		t.Fatalf("failed to create test ticket type: %v", err)
	}

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/submit", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.SubmitForApproval(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var logs []models.EventApprovalLog
	db.Where("event_id = ?", event.ID).Find(&logs)
	if len(logs) != 1 {
		t.Fatalf("expected exactly 1 approval log after submit, got %d", len(logs))
	}
	if logs[0].Action != models.ApprovalActionSubmitted {
		t.Errorf("expected log action 'submitted', got %s", logs[0].Action)
	}
	if logs[0].EventVersion != 1 {
		t.Errorf("expected log event_version 1, got %d", logs[0].EventVersion)
	}
}

func TestSubmitForApproval_WithoutTicketType_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia2", "panitia2@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/submit", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.SubmitForApproval(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 when event has no ticket type, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestSubmitForApproval_NotOwner_Forbidden(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia3", "panitia3@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	otherPanitia, _ := testutil.CreateTestUser(db, "panitia4", "panitia4@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	testutil.CreateTestTicketType(db, event.ID)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/submit", event.ID), otherPanitia.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.SubmitForApproval(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when a non-owner tries to submit the event, got %d", rr.Code)
	}
}

// ---- Approve / Reject / Request Revision ----

func TestApproveEvent_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia5", "panitia5@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	reviewer, _ := testutil.CreateTestUser(db, "superadmin1", "superadmin1@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPendingApproval)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/approve", event.ID), reviewer.ID, models.RoleSuperAdmin, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ApproveEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var logs []models.EventApprovalLog
	db.Where("event_id = ? AND action = ?", event.ID, models.ApprovalActionApproved).Find(&logs)
	if len(logs) != 1 {
		t.Fatalf("expected exactly 1 'approved' log entry, got %d", len(logs))
	}
}

func TestApproveEvent_SelfApproval_Forbidden(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "superadminorganizer", "superadminorganizer@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPendingApproval)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/approve", event.ID), organizer.ID, models.RoleSuperAdmin, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.ApproveEvent(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for self-approval attempt, got %d", rr.Code)
	}
}

func TestRejectEvent_RequiresNotes(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia6", "panitia6@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	reviewer, _ := testutil.CreateTestUser(db, "superadmin2", "superadmin2@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPendingApproval)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/reject", event.ID), reviewer.ID, models.RoleSuperAdmin, handlers.ReviewRequest{Notes: ""})
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.RejectEvent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when notes is missing for reject, got %d", rr.Code)
	}
}

func TestRejectEvent_ThenPublish_Fails(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia7", "panitia7@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	reviewer, _ := testutil.CreateTestUser(db, "superadmin3", "superadmin3@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPendingApproval)

	h := handlers.NewEventHandler(db)

	rejectReq := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/reject", event.ID), reviewer.ID, models.RoleSuperAdmin, handlers.ReviewRequest{Notes: "Tidak sesuai ketentuan"})
	rejectReq = withChiParam(rejectReq, "id", fmt.Sprintf("%d", event.ID))
	rejectRR := httptest.NewRecorder()
	h.RejectEvent(rejectRR, rejectReq)
	if rejectRR.Code != http.StatusOK {
		t.Fatalf("reject setup failed: expected 200, got %d", rejectRR.Code)
	}

	publishReq := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/publish", event.ID), reviewer.ID, models.RoleSuperAdmin, nil)
	publishReq = withChiParam(publishReq, "id", fmt.Sprintf("%d", event.ID))
	publishRR := httptest.NewRecorder()
	h.PublishEvent(publishRR, publishReq)

	if publishRR.Code != http.StatusConflict {
		t.Fatalf("expected 409 when trying to publish a rejected event, got %d", publishRR.Code)
	}
}

func TestRequestRevision_ThenResubmit_IncrementsVersion(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia8", "panitia8@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	reviewer, _ := testutil.CreateTestUser(db, "superadmin4", "superadmin4@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	testutil.CreateTestTicketType(db, event.ID)

	h := handlers.NewEventHandler(db)

	submitReq := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/submit", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	submitReq = withChiParam(submitReq, "id", fmt.Sprintf("%d", event.ID))
	h.SubmitForApproval(httptest.NewRecorder(), submitReq)

	revReq := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/request-revision", event.ID), reviewer.ID, models.RoleSuperAdmin, handlers.ReviewRequest{Notes: "Tolong lengkapi lokasi"})
	revReq = withChiParam(revReq, "id", fmt.Sprintf("%d", event.ID))
	revRR := httptest.NewRecorder()
	h.RequestRevision(revRR, revReq)
	if revRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for request-revision, got %d, body: %s", revRR.Code, revRR.Body.String())
	}

	submitReq2 := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/submit", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	submitReq2 = withChiParam(submitReq2, "id", fmt.Sprintf("%d", event.ID))
	submitRR2 := httptest.NewRecorder()
	h.SubmitForApproval(submitRR2, submitReq2)
	if submitRR2.Code != http.StatusOK {
		t.Fatalf("expected 200 for resubmit, got %d, body: %s", submitRR2.Code, submitRR2.Body.String())
	}

	var logs []models.EventApprovalLog
	db.Where("event_id = ?", event.ID).Order("log_id asc").Find(&logs)
	if len(logs) != 3 {
		t.Fatalf("expected 3 approval log entries (submitted, revision_requested, submitted), got %d", len(logs))
	}
	if logs[2].EventVersion != 2 {
		t.Errorf("expected 2nd submit log to have event_version 2, got %d", logs[2].EventVersion)
	}
}

// ---- Publish / Unpublish ----

func TestPublishEvent_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia9", "panitia9@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	superAdmin, _ := testutil.CreateTestUser(db, "superadmin5", "superadmin5@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusApproved)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/publish", event.ID), superAdmin.ID, models.RoleSuperAdmin, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.PublishEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

// TestPublishEvent_DoesNotWriteApprovalLog memvalidasi bahwa proses publikasi
// tidak mencatat data pada tabel riwayat persetujuan. Tabel event_approval_logs
// memiliki batasan validasi yang tidak menyertakan status 'published'.
// Pengujian ini mencegah terjadinya kegagalan integrasi apabila terdapat
// penambahan fungsi pencatatan secara sepihak di masa depan.
func TestPublishEvent_DoesNotWriteApprovalLog(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia10", "panitia10@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	superAdmin, _ := testutil.CreateTestUser(db, "superadmin6", "superadmin6@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusApproved)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/publish", event.ID), superAdmin.ID, models.RoleSuperAdmin, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	h.PublishEvent(httptest.NewRecorder(), req)

	var count int64
	db.Model(&models.EventApprovalLog{}).Where("event_id = ?", event.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 approval log rows after publish (action 'published' is not allowed by DB CHECK constraint), got %d", count)
	}
}

func TestUnpublishEvent_RevertsToDraft_DoesNotWriteApprovalLog(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia11", "panitia11@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	superAdmin, _ := testutil.CreateTestUser(db, "superadmin7", "superadmin7@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPublished)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/unpublish", event.ID), superAdmin.ID, models.RoleSuperAdmin, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.UnpublishEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}
	resp := testutil.DecodeData[models.Event](t, rr.Body.Bytes())
	if resp.Status != models.EventStatusDraft {
		t.Errorf("expected status draft after unpublish, got %s", resp.Status)
	}
	if resp.PublishedAt != nil {
		t.Error("expected published_at to be cleared after unpublish")
	}

	var count int64
	db.Model(&models.EventApprovalLog{}).Where("event_id = ?", event.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 approval log rows after unpublish, got %d", count)
	}
}

func TestPublishEvent_FromDraft_InvalidTransition(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia12", "panitia12@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	superAdmin, _ := testutil.CreateTestUser(db, "superadmin8", "superadmin8@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/publish", event.ID), superAdmin.ID, models.RoleSuperAdmin, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.PublishEvent(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 when publishing directly from draft, got %d", rr.Code)
	}
}

// ---- Get / List Approval Logs authorization ----

func TestGetByID_OwnerCanView(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia13", "panitia13@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodGet, fmt.Sprintf("/api/events/%d", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.GetByID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for owner viewing own event, got %d", rr.Code)
	}
}

func TestGetByID_OtherAdminPanitia_Forbidden(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia14", "panitia14@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	otherPanitia, _ := testutil.CreateTestUser(db, "panitia15", "panitia15@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodGet, fmt.Sprintf("/api/events/%d", event.ID), otherPanitia.ID, models.RoleAdminPanitia, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.GetByID(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a different admin_panitia viewing someone else's event, got %d", rr.Code)
	}
}

func TestGetByID_SuperAdminCanViewAnyEvent(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia16", "panitia16@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	superAdmin, _ := testutil.CreateTestUser(db, "superadmin9", "superadmin9@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodGet, fmt.Sprintf("/api/events/%d", event.ID), superAdmin.ID, models.RoleSuperAdmin, nil)
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.GetByID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for super_admin viewing any event, got %d", rr.Code)
	}
}

// TestGetByID_StafLapangan_CannotViewEvent memvalidasi batasan hak akses
// berdasarkan dokumen spesifikasi. Peran staf lapangan dikhususkan untuk
// fungsi operasional presensi sehingga tidak diizinkan mengakses detail
// kegiatan ataupun riwayat persetujuan.
func TestGetByID_StafLapangan_CannotViewEvent(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia17", "panitia17@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	staf, _ := testutil.CreateTestUser(db, "staf1", "staf1@eventgate.test", "Password123!", models.RoleStafLapangan, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)

	h := handlers.NewEventHandler(db)

	getReq := newEventRequest(http.MethodGet, fmt.Sprintf("/api/events/%d", event.ID), staf.ID, models.RoleStafLapangan, nil)
	getReq = withChiParam(getReq, "id", fmt.Sprintf("%d", event.ID))
	getRR := httptest.NewRecorder()
	h.GetByID(getRR, getReq)
	if getRR.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for staf_lapangan viewing event detail, got %d", getRR.Code)
	}

	logsReq := newEventRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/approval-logs", event.ID), staf.ID, models.RoleStafLapangan, nil)
	logsReq = withChiParam(logsReq, "id", fmt.Sprintf("%d", event.ID))
	logsRR := httptest.NewRecorder()
	h.ListApprovalLogs(logsRR, logsReq)
	if logsRR.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for staf_lapangan viewing approval logs, got %d", logsRR.Code)
	}
}

func TestListApprovalLogs_ReturnsFullHistory(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia18", "panitia18@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	reviewer, _ := testutil.CreateTestUser(db, "superadmin10", "superadmin10@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusDraft)
	testutil.CreateTestTicketType(db, event.ID)

	h := handlers.NewEventHandler(db)

	submitReq := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/submit", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	submitReq = withChiParam(submitReq, "id", fmt.Sprintf("%d", event.ID))
	h.SubmitForApproval(httptest.NewRecorder(), submitReq)

	approveReq := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/approve", event.ID), reviewer.ID, models.RoleSuperAdmin, nil)
	approveReq = withChiParam(approveReq, "id", fmt.Sprintf("%d", event.ID))
	h.ApproveEvent(httptest.NewRecorder(), approveReq)

	logsReq := newEventRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/approval-logs", event.ID), organizer.ID, models.RoleAdminPanitia, nil)
	logsReq = withChiParam(logsReq, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()
	h.ListApprovalLogs(rr, logsReq)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}
	logs := testutil.DecodeData[[]handlers.ApprovalLogResponse](t, rr.Body.Bytes())
	if len(logs) != 2 {
		t.Fatalf("expected 2 approval log entries (submitted, approved), got %d", len(logs))
	}
	if logs[0].Action != models.ApprovalActionSubmitted || logs[1].Action != models.ApprovalActionApproved {
		t.Errorf("expected log order [submitted, approved], got [%s, %s]", logs[0].Action, logs[1].Action)
	}
}

func TestRejectEvent_WithReasonField_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia19", "panitia19@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	reviewer, _ := testutil.CreateTestUser(db, "superadmin11", "superadmin11@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPendingApproval)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/reject", event.ID), reviewer.ID, models.RoleSuperAdmin, map[string]string{
		"reason": "Alasan penolakan melalui field reason",
	})
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.RejectEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for reject with reason field, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var logs []models.EventApprovalLog
	db.Where("event_id = ? AND action = ?", event.ID, models.ApprovalActionRejected).Find(&logs)
	if len(logs) != 1 {
		t.Fatalf("expected exactly 1 'rejected' log entry, got %d", len(logs))
	}
	if logs[0].Notes != "Alasan penolakan melalui field reason" {
		t.Errorf("expected log notes to match reason, got '%s'", logs[0].Notes)
	}
}

func TestRequestRevision_WithReasonField_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	organizer, _ := testutil.CreateTestUser(db, "panitia20", "panitia20@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	reviewer, _ := testutil.CreateTestUser(db, "superadmin12", "superadmin12@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	event, _ := testutil.CreateTestEvent(db, organizer.ID, models.EventStatusPendingApproval)

	h := handlers.NewEventHandler(db)
	req := newEventRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/request-revision", event.ID), reviewer.ID, models.RoleSuperAdmin, map[string]string{
		"reason": "Mohon lengkapi rincian kuota dan jadwal kegiatan pada deskripsi.",
	})
	req = withChiParam(req, "id", fmt.Sprintf("%d", event.ID))
	rr := httptest.NewRecorder()

	h.RequestRevision(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for request-revision with reason field, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var logs []models.EventApprovalLog
	db.Where("event_id = ? AND action = ?", event.ID, models.ApprovalActionRevisionRequested).Find(&logs)
	if len(logs) != 1 {
		t.Fatalf("expected exactly 1 'revision_requested' log entry, got %d", len(logs))
	}
	if logs[0].Notes != "Mohon lengkapi rincian kuota dan jadwal kegiatan pada deskripsi." {
		t.Errorf("expected log notes to match reason, got '%s'", logs[0].Notes)
	}
}
