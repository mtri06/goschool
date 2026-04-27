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
	createFn  func(s *model.NewStudent) error
	getByIDFn func(id int) (*model.StudentDetails, error)
	listFn    func(page, pageSize int, classID *int, name, email string) ([]model.StudentDetails, int, error)
	updateFn  func(id int, u *model.UpdateStudent) error
	deleteFn  func(id int) error
}

func (m *mockStudentSvc) CreateStudent(s *model.NewStudent) error {
	if m.createFn != nil {
		return m.createFn(s)
	}
	return nil
}
func (m *mockStudentSvc) GetStudentByID(id int) (*model.StudentDetails, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}
func (m *mockStudentSvc) ListStudents(page, pageSize int, classID *int, graduated *bool, name, email string) ([]model.StudentDetails, int, error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize, classID, name, email)
	}
	return []model.StudentDetails{}, 0, nil
}
func (m *mockStudentSvc) UpdateStudent(id int, u *model.UpdateStudent) error {
	if m.updateFn != nil {
		return m.updateFn(id, u)
	}
	return nil
}
func (m *mockStudentSvc) DeleteStudent(id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
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

func newValidUpdateStudent() *model.UpdateStudent {
	email := "john.updated@example.com"
	return &model.UpdateStudent{
		Email:         &email,
		Name:          "John Updated",
		DateOfBirth:   time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		AdmissionDate: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
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

// ---------------------------------------------------------------------------
// GetStudentByID
// ---------------------------------------------------------------------------

func TestStudentHandler_GetStudentByID_InvalidID(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodGet, "/students/abc", nil)
	req = withChiID(req, "abc")
	rr := httptest.NewRecorder()

	h.GetStudentByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudentByID_NotFound(t *testing.T) {
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		getByIDFn: func(id int) (*model.StudentDetails, error) { return nil, service.ErrNotFound },
	}
	req := httptest.NewRequest(http.MethodGet, "/students/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.GetStudentByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudentByID_Success(t *testing.T) {
	student := &model.StudentDetails{ID: 1, Name: "John Student"}
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		getByIDFn: func(id int) (*model.StudentDetails, error) { return student, nil },
	}
	req := httptest.NewRequest(http.MethodGet, "/students/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.GetStudentByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var resp model.StudentDetails
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != 1 || resp.Name != "John Student" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestStudentHandler_GetStudentByID_ServiceError(t *testing.T) {
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		getByIDFn: func(id int) (*model.StudentDetails, error) { return nil, errors.New("db error") },
	}
	req := httptest.NewRequest(http.MethodGet, "/students/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.GetStudentByID(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudentByID_PassesCorrectID(t *testing.T) {
	var capturedID *int
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		getByIDFn: func(id int) (*model.StudentDetails, error) {
			capturedID = &id
			return &model.StudentDetails{ID: id}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/students/42", nil)
	req = withChiID(req, "42")
	rr := httptest.NewRecorder()

	h.GetStudentByID(rr, req)

	if capturedID == nil || *capturedID != 42 {
		t.Errorf("expected service called with ID 42, got %v", capturedID)
	}
}

// ---------------------------------------------------------------------------
// GetStudents
// ---------------------------------------------------------------------------

func TestStudentHandler_GetStudents_Success(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	rr := httptest.NewRecorder()

	h.GetStudents(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudents_InvalidPage(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodGet, "/students?page=abc", nil)
	rr := httptest.NewRecorder()

	h.GetStudents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudents_InvalidPageSize(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodGet, "/students?pageSize=xyz", nil)
	rr := httptest.NewRecorder()

	h.GetStudents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudents_InvalidClassID(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodGet, "/students?classId=notanint", nil)
	rr := httptest.NewRecorder()

	h.GetStudents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudents_ServiceError(t *testing.T) {
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		listFn: func(page, pageSize int, classID *int, name, email string) ([]model.StudentDetails, int, error) {
			return nil, 0, errors.New("db error")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	rr := httptest.NewRecorder()

	h.GetStudents(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestStudentHandler_GetStudents_PassesQueryParams(t *testing.T) {
	var capturedClassID *int
	var capturedName, capturedEmail string
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		listFn: func(page, pageSize int, classID *int, name, email string) ([]model.StudentDetails, int, error) {
			capturedClassID = classID
			capturedName = name
			capturedEmail = email
			return []model.StudentDetails{}, 0, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/students?classId=5&year=2024&name=john&email=john%40example.com", nil)
	rr := httptest.NewRecorder()

	h.GetStudents(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedClassID == nil || *capturedClassID != 5 {
		t.Errorf("expected classID 5, got %v", capturedClassID)
	}
	if capturedName != "john" {
		t.Errorf("expected name 'john', got %q", capturedName)
	}
	if capturedEmail != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %q", capturedEmail)
	}
}

// ---------------------------------------------------------------------------
// UpdateStudent
// ---------------------------------------------------------------------------

func TestStudentHandler_UpdateStudent_InvalidID(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPut, "/students/abc", nil)
	req = withChiID(req, "abc")
	rr := httptest.NewRecorder()

	h.UpdateStudent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_UpdateStudent_InvalidBody(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPut, "/students/1", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.UpdateStudent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_UpdateStudent_Success(t *testing.T) {
	b, _ := json.Marshal(newValidUpdateStudent())
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPut, "/students/1", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.UpdateStudent(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestStudentHandler_UpdateStudent_NotFound(t *testing.T) {
	b, _ := json.Marshal(newValidUpdateStudent())
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		updateFn: func(id int, u *model.UpdateStudent) error { return service.ErrNotFound },
	}
	req := httptest.NewRequest(http.MethodPut, "/students/1", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.UpdateStudent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestStudentHandler_UpdateStudent_ServiceError(t *testing.T) {
	b, _ := json.Marshal(newValidUpdateStudent())
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		updateFn: func(id int, u *model.UpdateStudent) error { return errors.New("db error") },
	}
	req := httptest.NewRequest(http.MethodPut, "/students/1", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.UpdateStudent(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestStudentHandler_UpdateStudent_MissingRequiredField(t *testing.T) {
	h := newStudentHandlerWithMocks()
	requiredFields := []string{"name", "dateOfBirth", "gender", "admissionDate"}

	for _, field := range requiredFields {
		t.Run("missing "+field, func(t *testing.T) {
			body := map[string]any{
				"name":          "John Updated",
				"dateOfBirth":   "2008-01-01T00:00:00Z",
				"gender":        "male",
				"admissionDate": "2020-09-01T00:00:00Z",
			}
			delete(body, field)
			b, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPut, "/students/1", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = withChiID(req, "1")
			rr := httptest.NewRecorder()

			h.UpdateStudent(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for missing field %s, got %d", field, rr.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteStudent
// ---------------------------------------------------------------------------

func TestStudentHandler_DeleteStudent_InvalidID(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodDelete, "/students/abc", nil)
	req = withChiID(req, "abc")
	rr := httptest.NewRecorder()

	h.DeleteStudent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStudentHandler_DeleteStudent_Success(t *testing.T) {
	h := newStudentHandlerWithMocks()
	req := httptest.NewRequest(http.MethodDelete, "/students/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.DeleteStudent(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestStudentHandler_DeleteStudent_NotFound(t *testing.T) {
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		deleteFn: func(id int) error { return service.ErrNotFound },
	}
	req := httptest.NewRequest(http.MethodDelete, "/students/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.DeleteStudent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestStudentHandler_DeleteStudent_ServiceError(t *testing.T) {
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		deleteFn: func(id int) error { return errors.New("db error") },
	}
	req := httptest.NewRequest(http.MethodDelete, "/students/1", nil)
	req = withChiID(req, "1")
	rr := httptest.NewRecorder()

	h.DeleteStudent(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestStudentHandler_DeleteStudent_PassesCorrectID(t *testing.T) {
	var capturedID *int
	h := newStudentHandlerWithMocks()
	h.studentSvc = &mockStudentSvc{
		deleteFn: func(id int) error {
			capturedID = &id
			return nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/students/99", nil)
	req = withChiID(req, "99")
	rr := httptest.NewRecorder()

	h.DeleteStudent(rr, req)

	if capturedID == nil || *capturedID != 99 {
		t.Errorf("expected service called with ID 99, got %v", capturedID)
	}
}
