package service

import (
	"errors"
	"testing"
	"time"

	"goschool/pkg/model"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockStudentSvcUserRepo struct {
	createStudentFunc func(newStudent *model.NewStudent) error
}

func (m *mockStudentSvcUserRepo) CreateStudent(newStudent *model.NewStudent) error {
	if m.createStudentFunc != nil {
		return m.createStudentFunc(newStudent)
	}
	return nil
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
	return NewStudentService(&mockStudentSvcUserRepo{}, &mockStudentSvcUserSvc{})
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
	svc.studentRepo = &mockStudentSvcUserRepo{
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
		studentRepo *mockStudentSvcUserRepo
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
			studentRepo: &mockStudentSvcUserRepo{
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
