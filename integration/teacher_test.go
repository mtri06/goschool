package integration_test

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"goschool/pkg/model"
)

func TestTeacher_CRUD(t *testing.T) {
	cookies := loginAsAdmin(t)

	// ── 1. Create a teacher ──────────────────────────────────────────────────
	// Seed a subject first (required foreign key). Call the DB directly via
	// the app's own endpoint if available, or skip if not yet exposed.
	// For now we rely on the migration having seeded at least subject id=1.
	// If your migrations don't seed a subject, run an INSERT here via sqlx.

	newTeacher := model.NewTeacher{
		Username:      "t_integration",
		Password:      "Password1!",
		Name:          "Integration Teacher",
		DateOfBirth:   time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		SubjectID:     func() *int { id := 1; return &id }(),
		HireDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		WorkingStatus: "active",
	}

	createResp := requestJSON(t, http.MethodPost, "/teachers", newTeacher, withCookies(cookies))
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create teacher: expected 201, got %d", createResp.StatusCode)
	}

	// ── 2. List teachers — should include the new one ─────────────────────────
	listResp := requestJSON(t, http.MethodGet, "/teachers", nil, withCookies(cookies))
	var listBody struct {
		Teachers []model.TeacherDetails `json:"teachers"`
		Total    int                    `json:"total"`
	}
	decodeJSON(t, listResp, &listBody)

	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list teachers: expected 200, got %d", listResp.StatusCode)
	}
	if listBody.Total < 1 {
		t.Errorf("expected at least 1 teacher, got %d", listBody.Total)
	}

	// Find the created teacher ID
	var teacherID int
	for _, tc := range listBody.Teachers {
		if tc.Name == "Integration Teacher" {
			teacherID = tc.ID
			break
		}
	}
	if teacherID == 0 {
		t.Fatal("created teacher not found in list")
	}

	// ── 3. Get by ID ──────────────────────────────────────────────────────────
	getResp := requestJSON(t, http.MethodGet, "/teachers/"+itoa(teacherID), nil, withCookies(cookies))
	var got model.TeacherDetails
	decodeJSON(t, getResp, &got)

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get teacher: expected 200, got %d", getResp.StatusCode)
	}
	if got.Name != "Integration Teacher" {
		t.Errorf("expected name 'Integration Teacher', got %q", got.Name)
	}

	// ── 4. Update the teacher ─────────────────────────────────────────────────
	update := model.UpdateTeacher{
		Name:          "Updated Teacher",
		DateOfBirth:   time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		SubjectID:     func() *int { id := 1; return &id }(),
		HireDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		WorkingStatus: "on_leave",
	}

	updateResp := requestJSON(t, http.MethodPut, "/teachers/"+itoa(teacherID), update, withCookies(cookies))
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusNoContent {
		t.Fatalf("update teacher: expected 204, got %d", updateResp.StatusCode)
	}

	// ── 5. Verify update ──────────────────────────────────────────────────────
	getAfterUpdate := requestJSON(t, http.MethodGet, "/teachers/"+itoa(teacherID), nil, withCookies(cookies))
	var updated model.TeacherDetails
	decodeJSON(t, getAfterUpdate, &updated)

	if updated.Name != "Updated Teacher" {
		t.Errorf("expected updated name 'Updated Teacher', got %q", updated.Name)
	}
	if updated.WorkingStatus != "on_leave" {
		t.Errorf("expected working_status 'on_leave', got %q", updated.WorkingStatus)
	}

	// ── 6. Delete the teacher ─────────────────────────────────────────────────
	deleteResp := requestJSON(t, http.MethodDelete, "/teachers/"+itoa(teacherID), nil, withCookies(cookies))
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete teacher: expected 204, got %d", deleteResp.StatusCode)
	}

	// ── 7. Get after delete — should be 404 ───────────────────────────────────
	getAfterDelete := requestJSON(t, http.MethodGet, "/teachers/"+itoa(teacherID), nil, withCookies(cookies))
	defer getAfterDelete.Body.Close()

	if getAfterDelete.StatusCode != http.StatusNotFound {
		t.Fatalf("get deleted teacher: expected 404, got %d", getAfterDelete.StatusCode)
	}
}

func TestTeacher_Unauthorized(t *testing.T) {
	resp := requestJSON(t, http.MethodGet, "/teachers", nil, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", resp.StatusCode)
	}
}

func TestTeacher_GetByID_NotFound(t *testing.T) {
	cookies := loginAsAdmin(t)

	resp := requestJSON(t, http.MethodGet, "/teachers/99999", nil, withCookies(cookies))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent teacher, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func itoa(n int) string {
	return strconv.Itoa(n)
}
