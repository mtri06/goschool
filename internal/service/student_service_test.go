package service

import (
	"errors"
	"testing"
	"time"

	"goschool/internal/repository"
	"goschool/pkg/model"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockStudentUserRepo struct {
	emailExistsFn func(email string, excludeIDs ...int) (bool, error)
}

func (m *mockStudentUserRepo) EmailExists(email string, excludeIDs ...int) (bool, error) {
	if m.emailExistsFn != nil {
		return m.emailExistsFn(email, excludeIDs...)
	}
	return false, nil
}

type mockStudentRepo struct {
	createStudentFunc func(newStudent *model.NewStudent) error
	getByIDFn         func(id int) (*model.StudentDetails, error)
	studentExistsFn   func(id int) (bool, error)
	updateStudentFn   func(id int, u *model.UpdateStudent) error
	deleteStudentFn   func(id int) error
	listStudentsFn    func(p *repository.Pagination, uf repository.Filters, sf repository.Filters) ([]model.StudentDetails, int, error)
}

func (m *mockStudentRepo) CreateStudent(newStudent *model.NewStudent) error {
	if m.createStudentFunc != nil {
		return m.createStudentFunc(newStudent)
	}
	return nil
}
func (m *mockStudentRepo) GetStudentByID(id int) (*model.StudentDetails, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}
func (m *mockStudentRepo) StudentExists(id int) (bool, error) {
	if m.studentExistsFn != nil {
		return m.studentExistsFn(id)
	}
	return true, nil
}
func (m *mockStudentRepo) UpdateStudent(id int, u *model.UpdateStudent) error {
	if m.updateStudentFn != nil {
		return m.updateStudentFn(id, u)
	}
	return nil
}
func (m *mockStudentRepo) DeleteStudent(id int) error {
	if m.deleteStudentFn != nil {
		return m.deleteStudentFn(id)
	}
	return nil
}
func (m *mockStudentRepo) ListStudents(p *repository.Pagination, uf repository.Filters, ef repository.Filters) ([]model.StudentDetails, int, error) {
	if m.listStudentsFn != nil {
		return m.listStudentsFn(p, uf, ef)
	}
	return []model.StudentDetails{}, 0, nil
}

type mockStudentClassRepo struct {
	classExistsFn func(id int) (bool, error)
}

func (m *mockStudentClassRepo) ClassExists(id int) (bool, error) {
	if m.classExistsFn != nil {
		return m.classExistsFn(id)
	}
	return true, nil
}

type mockStudentSvcUserSvc struct {
	validateUserFunc func(user *model.User) error
}

func (m *mockStudentSvcUserSvc) validateUser(user *model.User) error {
	if m.validateUserFunc != nil {
		return m.validateUserFunc(user)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newStudentServiceWithMocks() *StudentService {
	return NewStudentService(&mockStudentUserRepo{}, &mockStudentRepo{}, &mockStudentClassRepo{}, &mockStudentSvcUserSvc{})
}

func validNewStudent() *model.NewStudent {
	mail := "test.student@example.com"
	return &model.NewStudent{
		Username:      "teststudent",
		Password:      "TestPass123!",
		Email:         &mail,
		Name:          "John Student",
		DateOfBirth:   time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		AdmissionDate: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
	}
}

func validUpdateStudent() *model.UpdateStudent {
	mail := "updated.student@example.com"
	return &model.UpdateStudent{
		Email:         &mail,
		Name:          "Updated Student",
		DateOfBirth:   time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        "female",
		AdmissionDate: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ---------------------------------------------------------------------------
// TestCreateStudent — focused tests
// ---------------------------------------------------------------------------

func TestStudentService_CreateStudent_MustCallValidateUser(t *testing.T) {
	called := false

	svc := newStudentServiceWithMocks()
	svc.userSvc = &mockStudentSvcUserSvc{
		validateUserFunc: func(u *model.User) error {
			called = true
			return nil
		},
	}

	if err := svc.CreateStudent(validNewStudent()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected validateUser to be called, but it was not")
	}
}

func TestStudentService_CreateStudent_PasswordMustBeHashed(t *testing.T) {
	var captured *model.NewStudent
	newStudent := validNewStudent()
	plain := newStudent.Password

	svc := newStudentServiceWithMocks()
	svc.studentRepo = &mockStudentRepo{
		createStudentFunc: func(s *model.NewStudent) error {
			captured = s
			return nil
		},
	}

	if err := svc.CreateStudent(newStudent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected repo.CreateStudent to be called")
	}
	if captured.Password == plain {
		t.Error("expected password to be hashed, but it was stored in plain text")
	}
}

func TestStudentService_CreateStudent(t *testing.T) {
	dbErr := errors.New("db error")

	tests := []struct {
		name        string
		input       *model.NewStudent
		studentRepo *mockStudentRepo
		userSvc     *mockStudentSvcUserSvc
		wantErr     error
	}{
		{
			name:  "valid input",
			input: validNewStudent(),
		},
		{
			name: "nil email is allowed",
			input: func() *model.NewStudent {
				s := validNewStudent()
				s.Email = nil
				return s
			}(),
		},
		{
			name:  "validateUser returns validation error",
			input: validNewStudent(),
			userSvc: &mockStudentSvcUserSvc{
				validateUserFunc: func(u *model.User) error { return ErrValidationFailed },
			},
			wantErr: ErrValidationFailed,
		},
		{
			name:  "validateUser returns db error",
			input: validNewStudent(),
			userSvc: &mockStudentSvcUserSvc{
				validateUserFunc: func(u *model.User) error { return dbErr },
			},
			wantErr: dbErr,
		},
		{
			name:  "repo returns db error",
			input: validNewStudent(),
			studentRepo: &mockStudentRepo{
				createStudentFunc: func(s *model.NewStudent) error { return dbErr },
			},
			wantErr: dbErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newStudentServiceWithMocks()
			if tc.studentRepo != nil {
				svc.studentRepo = tc.studentRepo
			}
			if tc.userSvc != nil {
				svc.userSvc = tc.userSvc
			}

			err := svc.CreateStudent(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetStudentByID
// ---------------------------------------------------------------------------

func TestStudentService_GetStudentByID(t *testing.T) {
	dbErr := errors.New("db error")
	student := &model.StudentDetails{ID: 1, Name: "John Student"}

	tests := []struct {
		name        string
		studentRepo *mockStudentRepo
		wantStudent *model.StudentDetails
		wantErr     error
	}{
		{
			name:        "found",
			studentRepo: &mockStudentRepo{getByIDFn: func(id int) (*model.StudentDetails, error) { return student, nil }},
			wantStudent: student,
		},
		{
			name:        "not found",
			studentRepo: &mockStudentRepo{getByIDFn: func(id int) (*model.StudentDetails, error) { return nil, nil }},
			wantErr:     ErrNotFound,
		},
		{
			name:        "repo error",
			studentRepo: &mockStudentRepo{getByIDFn: func(id int) (*model.StudentDetails, error) { return nil, dbErr }},
			wantErr:     dbErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newStudentServiceWithMocks()
			svc.studentRepo = tc.studentRepo

			got, err := svc.GetStudentByID(1)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
			if tc.wantStudent != nil && got != tc.wantStudent {
				t.Errorf("got %v, want %v", got, tc.wantStudent)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUpdateStudent
// ---------------------------------------------------------------------------

func TestStudentService_UpdateStudent(t *testing.T) {
	dbErr := errors.New("db error")

	tests := []struct {
		name        string
		input       *model.UpdateStudent
		studentRepo *mockStudentRepo
		userRepo    *mockStudentUserRepo
		wantErr     error
	}{
		{
			name:  "success",
			input: validUpdateStudent(),
		},
		{
			name: "invalid gender",
			input: func() *model.UpdateStudent {
				u := validUpdateStudent()
				u.Gender = "invalid"
				return u
			}(),
			wantErr: ErrValidationFailed,
		},
		{
			name:        "student not found",
			input:       validUpdateStudent(),
			studentRepo: &mockStudentRepo{studentExistsFn: func(id int) (bool, error) { return false, nil }},
			wantErr:     ErrNotFound,
		},
		{
			name:        "student exists check fails",
			input:       validUpdateStudent(),
			studentRepo: &mockStudentRepo{studentExistsFn: func(id int) (bool, error) { return false, dbErr }},
			wantErr:     dbErr,
		},
		{
			name:     "email already exists",
			input:    validUpdateStudent(),
			userRepo: &mockStudentUserRepo{emailExistsFn: func(email string, excludeIDs ...int) (bool, error) { return true, nil }},
			wantErr:  ErrValidationFailed,
		},
		{
			name:     "email check fails",
			input:    validUpdateStudent(),
			userRepo: &mockStudentUserRepo{emailExistsFn: func(email string, excludeIDs ...int) (bool, error) { return false, dbErr }},
			wantErr:  dbErr,
		},
		{
			name:        "repo update fails",
			input:       validUpdateStudent(),
			studentRepo: &mockStudentRepo{updateStudentFn: func(id int, u *model.UpdateStudent) error { return dbErr }},
			wantErr:     dbErr,
		},
		{
			name: "nil email skips email check",
			input: func() *model.UpdateStudent {
				u := validUpdateStudent()
				u.Email = nil
				return u
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newStudentServiceWithMocks()
			if tc.studentRepo != nil {
				svc.studentRepo = tc.studentRepo
			}
			if tc.userRepo != nil {
				svc.userRepo = tc.userRepo
			}

			err := svc.UpdateStudent(1, tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDeleteStudent
// ---------------------------------------------------------------------------

func TestStudentService_DeleteStudent(t *testing.T) {
	dbErr := errors.New("db error")

	tests := []struct {
		name        string
		studentRepo *mockStudentRepo
		wantErr     error
	}{
		{
			name: "success",
		},
		{
			name:        "student not found",
			studentRepo: &mockStudentRepo{studentExistsFn: func(id int) (bool, error) { return false, nil }},
			wantErr:     ErrNotFound,
		},
		{
			name:        "exists check fails",
			studentRepo: &mockStudentRepo{studentExistsFn: func(id int) (bool, error) { return false, dbErr }},
			wantErr:     dbErr,
		},
		{
			name:        "delete fails",
			studentRepo: &mockStudentRepo{deleteStudentFn: func(id int) error { return dbErr }},
			wantErr:     dbErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newStudentServiceWithMocks()
			if tc.studentRepo != nil {
				svc.studentRepo = tc.studentRepo
			}

			err := svc.DeleteStudent(1)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestListStudents
// ---------------------------------------------------------------------------

func TestStudentService_ListStudents(t *testing.T) {
	dbErr := errors.New("db error")
	students := []model.StudentDetails{{ID: 1, Name: "John"}}

	tests := []struct {
		name        string
		studentRepo *mockStudentRepo
		wantTotal   int
		wantLen     int
		wantErr     error
	}{
		{
			name: "success",
			studentRepo: &mockStudentRepo{
				listStudentsFn: func(p *repository.Pagination, uf repository.Filters, ef repository.Filters) ([]model.StudentDetails, int, error) {
					return students, 1, nil
				},
			},
			wantTotal: 1,
			wantLen:   1,
		},
		{
			name: "repo error",
			studentRepo: &mockStudentRepo{
				listStudentsFn: func(p *repository.Pagination, uf repository.Filters, ef repository.Filters) ([]model.StudentDetails, int, error) {
					return nil, 0, dbErr
				},
			},
			wantErr: dbErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newStudentServiceWithMocks()
			svc.studentRepo = tc.studentRepo

			got, total, err := svc.ListStudents(1, 20, nil, nil, "", "")
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
			if total != tc.wantTotal {
				t.Errorf("got total %d, want %d", total, tc.wantTotal)
			}
			if len(got) != tc.wantLen {
				t.Errorf("got %d students, want %d", len(got), tc.wantLen)
			}
		})
	}
}
