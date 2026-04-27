package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goschool/internal/service"
	"goschool/pkg/model"

	"github.com/go-chi/chi/v5"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockTeacherSvc struct {
	createFn  func(t *model.NewTeacher) error
	getByIDFn func(id int) (*model.TeacherDetails, error)
	updateFn  func(id int, u *model.UpdateTeacher) error
	listFn    func(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error)
	deleteFn  func(id int) error
}

func (m *mockTeacherSvc) CreateTeacher(t *model.NewTeacher) error {
	if m.createFn != nil {
		return m.createFn(t)
	}
	return nil
}
func (m *mockTeacherSvc) GetTeacherByID(id int) (*model.TeacherDetails, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}
func (m *mockTeacherSvc) UpdateTeacher(id int, u *model.UpdateTeacher) error {
	if m.updateFn != nil {
		return m.updateFn(id, u)
	}
	return nil
}
func (m *mockTeacherSvc) ListTeachers(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize, name, email, workingStatus)
	}
	return []model.TeacherDetails{}, 0, nil
}
func (m *mockTeacherSvc) DeleteTeacher(id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTeacherHandlerWithMocks() *TeacherHandler {
	return NewTeacherHandler(&mockTeacherSvc{}, NewErrorMap())
}

func withChiID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ---------------------------------------------------------------------------
// GetTeacherByID
// ---------------------------------------------------------------------------

func TestTeacherHandler_GetTeacherByID_InvalidID(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	req := httptest.NewRequest(http.MethodGet, "/teachers/abc", nil)
	req = withChiID(req, "abc")
	rr := httptest.NewRecorder()

	h.GetTeacherByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTeacherHandler_GetTeacherByID_NotFound(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		getByIDFn: func(id int) (*model.TeacherDetails, error) {
			return nil, service.ErrNotFound
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/teachers/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.GetTeacherByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestTeacherHandler_GetTeacherByID_Success(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	teacher := &model.TeacherDetails{ID: 1, Name: "John Doe"}
	h.teacherSvc = &mockTeacherSvc{
		getByIDFn: func(id int) (*model.TeacherDetails, error) { return teacher, nil },
	}
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
	if resp.ID != 1 || resp.Name != "John Doe" {
		t.Errorf("expected ID 1 and name John Doe, got ID %d and name %s", resp.ID, resp.Name)
	}
}

func TestTeacherHandler_GetTeacherByID_ServiceUnknownError(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		getByIDFn: func(id int) (*model.TeacherDetails, error) {
			return nil, errors.New("db error")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/teachers/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.GetTeacherByID(rr, req)

	if !is500Error(rr) {
		t.Errorf("expected 500 with unknown internal error payload, got %d", rr.Code)
	}
}

func TestTeacherHandler_GetTeacherByID_MustPassCorrectIDToService(t *testing.T) {
	var capturedID *int
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		getByIDFn: func(id int) (*model.TeacherDetails, error) {
			capturedID = &id
			return &model.TeacherDetails{ID: id, Name: "John Doe"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/teachers/42", nil)
	req = withChiID(req, "42")
	rr := httptest.NewRecorder()

	h.GetTeacherByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedID == nil || *capturedID != 42 {
		t.Errorf("expected GetTeacherByID to be called with ID 42, got %v", capturedID)
	}
}

// ---------------------------------------------------------------------------
// DeleteTeacher
// ---------------------------------------------------------------------------

func TestTeacherHandler_DeleteTeacher_InvalidID(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	req := httptest.NewRequest(http.MethodDelete, "/teachers/xyz", nil)
	req = withChiID(req, "xyz")
	rr := httptest.NewRecorder()

	h.DeleteTeacher(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTeacherHandler_DeleteTeacher_Success(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	req := httptest.NewRequest(http.MethodDelete, "/teachers/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.DeleteTeacher(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestTeacherHandler_DeleteTeacher_ServiceUnknownError(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		deleteFn: func(id int) error {
			return errors.New("unknown error")
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/teachers/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.DeleteTeacher(rr, req)

	if !is500Error(rr) {
		t.Errorf("expected 500 with unknown internal error payload, got %v", rr)
	}
}

func TestTeacherHandler_DeleteTeacher_MustPassCorrectIDToService(t *testing.T) {
	var capturedID *int
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		deleteFn: func(id int) error {
			capturedID = &id
			return nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/teachers/99", nil)
	req = withChiID(req, "99")
	rr := httptest.NewRecorder()

	h.DeleteTeacher(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if capturedID == nil || *capturedID != 99 {
		t.Errorf("expected DeleteTeacher to be called with ID 99, got %v", capturedID)
	}
}

func TestTeacherHandler_DeleteTeacher_NotFound(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		deleteFn: func(id int) error {
			return service.ErrNotFound
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/teachers/123", nil)
	req = withChiID(req, "123")
	rr := httptest.NewRecorder()

	h.DeleteTeacher(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// CreateTeacher
// ---------------------------------------------------------------------------

func newValidNewTeacher() *model.NewTeacher {
	return &model.NewTeacher{
		Username:      "jdoe",
		Password:      "Password1!",
		Name:          "John Doe",
		DateOfBirth:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		SubjectID:     1,
		HireDate:      time.Now(),
		WorkingStatus: "active",
	}
}

func TestTeacherHandler_CreateTeacher_InvalidBody(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPost, "/teachers", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTeacher(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTeacherHandler_CreateTeacher_Success(t *testing.T) {
	body := newValidNewTeacher()
	b, _ := json.Marshal(body)

	h := newTeacherHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPost, "/teachers", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTeacher(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestTeacherHandler_CreateTeacher_ServiceUnknownError(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		createFn: func(t *model.NewTeacher) error {
			return errors.New("db error")
		},
	}
	body := newValidNewTeacher()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/teachers", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTeacher(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestTeacherHandler_CreateTeacher_MissingRequiredField(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	requiredFields := []string{
		"username", "password", "name", "dateOfBirth", "gender", "subjectID",
		"hireDate", "workingStatus",
	}

	for _, field := range requiredFields {
		t.Run("missing "+field, func(t *testing.T) {
			body := map[string]any{
				"username":      "jdoe",
				"password":      "Password1!",
				"name":          "John Doe",
				"dateOfBirth":   "1990-01-01T00:00:00Z",
				"gender":        "male",
				"subjectID":     1,
				"hireDate":      time.Now().Format(time.RFC3339),
				"workingStatus": "active",
			}
			delete(body, field)
			b, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/teachers", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.CreateTeacher(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for missing field %s, got %d", field, rr.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetTeachers
// ---------------------------------------------------------------------------

func TestTeacherHandler_GetTeachers_Success(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		listFn: func(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error) {
			return []model.TeacherDetails{{ID: 1, Name: "Alice"}}, 1, nil
		},
	}
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

func TestTeacherHandler_GetTeachers_ServiceError(t *testing.T) {
	h := newTeacherHandlerWithMocks()
	h.teacherSvc = &mockTeacherSvc{
		listFn: func(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error) {
			return nil, 0, errors.New("db error")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
	rr := httptest.NewRecorder()

	h.GetTeachers(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
