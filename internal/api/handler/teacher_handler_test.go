package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goschool/internal/api/handler"
	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/chi/v5"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockTeacherSvc struct {
	createFn  func(t *model.NewTeacher) error
	getByIDFn func(id int64) (*model.TeacherDetails, error)
	updateFn  func(id int64, u *model.UpdateTeacher) error
	listFn    func(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error)
	deleteFn  func(id int64) error
}

func (m *mockTeacherSvc) CreateTeacher(t *model.NewTeacher) error {
	if m.createFn != nil {
		return m.createFn(t)
	}
	return nil
}
func (m *mockTeacherSvc) GetTeacherByID(id int64) (*model.TeacherDetails, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}
func (m *mockTeacherSvc) UpdateTeacher(id int64, u *model.UpdateTeacher) error {
	if m.updateFn != nil {
		return m.updateFn(id, u)
	}
	return nil
}
func (m *mockTeacherSvc) ListTeachers(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize, name, email)
	}
	return []model.TeacherDetails{}, 0, nil
}
func (m *mockTeacherSvc) DeleteTeacher(id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

var testErrMap = httpx.APIErrorMap{
	httpx.ErrInvalidBody:  httpx.ErrBadRequest,
	httpx.ErrInvalidQuery: httpx.ErrBadRequest,
}

func newHandler(svc *mockTeacherSvc) *handler.TeacherHandler {
	return handler.NewTeacherHandler(svc, testErrMap)
}

func withChiID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ---------------------------------------------------------------------------
// GetTeacherByID
// ---------------------------------------------------------------------------

func TestGetTeacherByID_InvalidID(t *testing.T) {
	h := newHandler(&mockTeacherSvc{})
	req := httptest.NewRequest(http.MethodGet, "/teachers/abc", nil)
	req = withChiID(req, "abc")
	rr := httptest.NewRecorder()

	h.GetTeacherByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestGetTeacherByID_NotFound(t *testing.T) {
	svc := &mockTeacherSvc{
		getByIDFn: func(id int64) (*model.TeacherDetails, error) {
			return nil, httpx.ErrNotFound
		},
	}
	h := newHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/teachers/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.GetTeacherByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestGetTeacherByID_Success(t *testing.T) {
	teacher := &model.TeacherDetails{ID: 1, Name: "John Doe"}
	svc := &mockTeacherSvc{
		getByIDFn: func(id int64) (*model.TeacherDetails, error) { return teacher, nil },
	}
	h := newHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/teachers/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.GetTeacherByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var resp model.TeacherDetails
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Name != "John Doe" {
		t.Errorf("expected name John Doe, got %s", resp.Name)
	}
}

// ---------------------------------------------------------------------------
// DeleteTeacher
// ---------------------------------------------------------------------------

func TestDeleteTeacher_InvalidID(t *testing.T) {
	h := newHandler(&mockTeacherSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/teachers/xyz", nil)
	req = withChiID(req, "xyz")
	rr := httptest.NewRecorder()

	h.DeleteTeacher(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDeleteTeacher_Success(t *testing.T) {
	h := newHandler(&mockTeacherSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/teachers/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.DeleteTeacher(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// CreateTeacher
// ---------------------------------------------------------------------------

func TestCreateTeacher_InvalidBody(t *testing.T) {
	h := newHandler(&mockTeacherSvc{})
	req := httptest.NewRequest(http.MethodPost, "/teachers", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTeacher(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCreateTeacher_Success(t *testing.T) {
	body := model.NewTeacher{
		Username:      "jdoe",
		Password:      "Password1!",
		Name:          "John Doe",
		DateOfBirth:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		SubjectID:     1,
		HireDate:      time.Now(),
		WorkingStatus: "active",
	}
	b, _ := json.Marshal(body)

	h := newHandler(&mockTeacherSvc{})
	req := httptest.NewRequest(http.MethodPost, "/teachers", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTeacher(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// GetTeachers
// ---------------------------------------------------------------------------

func TestGetTeachers_Success(t *testing.T) {
	svc := &mockTeacherSvc{
		listFn: func(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error) {
			return []model.TeacherDetails{{ID: 1, Name: "Alice"}}, 1, nil
		},
	}
	h := newHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/teachers?page=1&pageSize=10", nil)
	rr := httptest.NewRecorder()

	h.GetTeachers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

func TestGetTeachers_ServiceError(t *testing.T) {
	svc := &mockTeacherSvc{
		listFn: func(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error) {
			return nil, 0, errors.New("db error")
		},
	}
	h := newHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
	rr := httptest.NewRecorder()

	h.GetTeachers(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
