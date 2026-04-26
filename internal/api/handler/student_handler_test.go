package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goschool/internal/service"
	"goschool/pkg/model"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockStudentSvc struct {
	createFn func(s *model.NewStudent) error
}

func (m *mockStudentSvc) CreateStudent(s *model.NewStudent) error {
	if m.createFn != nil {
		return m.createFn(s)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newStudentHandlerWithMocks() *StudentHandler {
	return NewStudentHandler(&mockStudentSvc{}, NewErrorMap())
}

func newValidNewStudent() *model.NewStudent {
	email := "john.student@example.com"
	return &model.NewStudent{
		Username:      "jstudent",
		Password:      "Password1!",
		Name:          "John Student",
		DateOfBirth:   time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		AdmissionDate: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		Email:         &email,
	}
}

// ---------------------------------------------------------------------------
// CreateStudent
// ---------------------------------------------------------------------------

func TestStudentHandler_CreateStudent_InvalidBody(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateStudent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_CreateStudent_Success(t *testing.T) {
	b, _ := json.Marshal(newValidNewStudent())
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateStudent(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestStudentHandler_CreateStudent_ServiceUnknownError(t *testing.T) {
	b, _ := json.Marshal(newValidNewStudent())
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		createFn: func(s *model.NewStudent) error { return errors.New("db error") },
	}
	req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateStudent(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestStudentHandler_CreateStudent_ServiceValidationError(t *testing.T) {
	b, _ := json.Marshal(newValidNewStudent())
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		createFn: func(s *model.NewStudent) error { return service.ErrValidationFailed },
	}
	req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateStudent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_CreateStudent_MissingRequiredField(t *testing.T) {
	h := newStudentHandlerWithMocks()
	requiredFields := []string{
		"username", "password", "name", "dateOfBirth", "gender", "admissionDate",
	}

	for _, field := range requiredFields {
		t.Run("missing "+field, func(t *testing.T) {
			body := map[string]any{
				"username":      "jstudent",
				"password":      "Password1!",
				"name":          "John Student",
				"dateOfBirth":   "2008-01-01T00:00:00Z",
				"gender":        "male",
				"admissionDate": "2020-09-01T00:00:00Z",
			}
			delete(body, field)
			b, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.CreateStudent(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for missing field %s, got %d", field, rr.Code)
			}
		})
	}
}
