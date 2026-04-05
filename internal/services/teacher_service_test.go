package services

import (
	"errors"
	"testing"
	"time"

	"goschool/pkg/constant"
	"goschool/pkg/model"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockTeacherUserRepo struct {
	emailExistsFn func(email string) (bool, error)
}

func (m *mockTeacherUserRepo) EmailExists(email string) (bool, error) {
	if m.emailExistsFn != nil {
		return m.emailExistsFn(email)
	}
	return false, nil
}

type mockTeacherRepo struct {
	createFn        func(t *model.NewTeacher) error
	getByIDFn       func(id int64) (*model.TeacherDetails, error)
	teacherExistsFn func(id int64) (bool, error)
	updateFn        func(id int64, u *model.UpdateTeacher) error
	deleteFn        func(id int64) error
	listFn          func(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error)
}

func (m *mockTeacherRepo) CreateTeacher(t *model.NewTeacher) error {
	if m.createFn != nil {
		return m.createFn(t)
	}
	return nil
}
func (m *mockTeacherRepo) GetTeacherByID(id int64) (*model.TeacherDetails, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}
func (m *mockTeacherRepo) TeacherExists(id int64) (bool, error) {
	if m.teacherExistsFn != nil {
		return m.teacherExistsFn(id)
	}
	return false, nil
}
func (m *mockTeacherRepo) UpdateTeacher(id int64, u *model.UpdateTeacher) error {
	if m.updateFn != nil {
		return m.updateFn(id, u)
	}
	return nil
}
func (m *mockTeacherRepo) DeleteTeacher(id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}
func (m *mockTeacherRepo) ListTeachers(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize, name, email)
	}
	return nil, 0, nil
}

type mockSubjectRepo struct {
	existsFn func(id int64) (bool, error)
}

func (m *mockSubjectRepo) Exists(id int64) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(id)
	}
	return false, nil
}

type mockUserSvcForTeacher struct {
	validateUserFn func(user *model.User) error
}

func (m *mockUserSvcForTeacher) validateUser(user *model.User) error {
	if m.validateUserFn != nil {
		return m.validateUserFn(user)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func validNewTeacher() *model.NewTeacher {
	return &model.NewTeacher{
		Username:      "jdoe",
		Password:      "Password1!",
		Name:          "John Doe",
		DateOfBirth:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        constant.GenderMale,
		SubjectID:     1,
		HireDate:      time.Now(),
		WorkingStatus: constant.WorkingStatusActive,
	}
}

func validUpdateTeacher() *model.UpdateTeacher {
	return &model.UpdateTeacher{
		Name:          "Jane Doe",
		DateOfBirth:   time.Date(1991, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        constant.GenderFemale,
		SubjectID:     1,
		HireDate:      time.Now(),
		WorkingStatus: constant.WorkingStatusActive,
	}
}

func newTeacherSvc(userRepo TeacherSvcUserRepo, teacherRepo UserSvcTeacherRepo, subjectRepo TeacherSvcSubjectRepo, userSvc TeacherSvcUserSvc) *TeacherService {
	return NewTeacherService(userRepo, teacherRepo, subjectRepo, userSvc)
}

// ---------------------------------------------------------------------------
// TestCreateTeacher
// ---------------------------------------------------------------------------

func TestTeacherService_CreateTeacher(t *testing.T) {
	anyErr := errors.New("db error")

	tests := []struct {
		name        string
		input       *model.NewTeacher
		userRepo    *mockTeacherUserRepo
		teacherRepo *mockTeacherRepo
		subjectRepo *mockSubjectRepo
		userSvc     *mockUserSvcForTeacher
		wantErr     error
	}{
		{
			name:        "success",
			input:       validNewTeacher(),
			userRepo:    &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{},
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return true, nil }},
			userSvc:     &mockUserSvcForTeacher{},
			wantErr:     nil,
		},
		{
			name:  "invalid working status",
			input: func() *model.NewTeacher { t := validNewTeacher(); t.WorkingStatus = "unknown"; return t }(),
			userRepo:    &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{},
			subjectRepo: &mockSubjectRepo{},
			userSvc:     &mockUserSvcForTeacher{},
			wantErr:     ErrValidationFailed,
		},
		{
			name:        "subject not found",
			input:       validNewTeacher(),
			userRepo:    &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{},
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return false, nil }},
			userSvc:     &mockUserSvcForTeacher{},
			wantErr:     ErrNotFound,
		},
		{
			name:        "subject repo error",
			input:       validNewTeacher(),
			userRepo:    &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{},
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return false, anyErr }},
			userSvc:     &mockUserSvcForTeacher{},
			wantErr:     anyErr,
		},
		{
			name:        "validateUser returns error",
			input:       validNewTeacher(),
			userRepo:    &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{},
			subjectRepo: &mockSubjectRepo{},
			userSvc:     &mockUserSvcForTeacher{validateUserFn: func(u *model.User) error { return ErrValidationFailed }},
			wantErr:     ErrValidationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTeacherSvc(tc.userRepo, tc.teacherRepo, tc.subjectRepo, tc.userSvc)
			err := svc.CreateTeacher(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetTeacherByID
// ---------------------------------------------------------------------------

func TestTeacherService_GetTeacherByID(t *testing.T) {
	teacher := &model.TeacherDetails{ID: 1, Name: "John Doe"}
	anyErr := errors.New("db error")

	tests := []struct {
		name        string
		id          int64
		teacherRepo *mockTeacherRepo
		wantErr     error
		wantNil     bool
	}{
		{
			name:        "success",
			id:          1,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int64) (*model.TeacherDetails, error) { return teacher, nil }},
			wantErr:     nil,
		},
		{
			name:        "teacher not found",
			id:          99,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int64) (*model.TeacherDetails, error) { return nil, nil }},
			wantErr:     ErrNotFound,
			wantNil:     true,
		},
		{
			name:        "repo error",
			id:          1,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int64) (*model.TeacherDetails, error) { return nil, anyErr }},
			wantErr:     anyErr,
			wantNil:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTeacherSvc(&mockTeacherUserRepo{}, tc.teacherRepo, &mockSubjectRepo{}, &mockUserSvcForTeacher{})
			got, err := svc.GetTeacherByID(tc.id)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
			if tc.wantNil && got != nil {
				t.Errorf("expected nil result, got %+v", got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDeleteTeacher
// ---------------------------------------------------------------------------

func TestTeacherService_DeleteTeacher(t *testing.T) {
	anyErr := errors.New("db error")

	tests := []struct {
		name        string
		id          int64
		teacherRepo *mockTeacherRepo
		wantErr     error
	}{
		{
			name: "success",
			id:   1,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return true, nil },
				deleteFn:        func(id int64) error { return nil },
			},
			wantErr: nil,
		},
		{
			name: "teacher not found",
			id:   99,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return false, nil },
			},
			wantErr: ErrNotFound,
		},
		{
			name: "exists check error",
			id:   1,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return false, anyErr },
			},
			wantErr: anyErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTeacherSvc(&mockTeacherUserRepo{}, tc.teacherRepo, &mockSubjectRepo{}, &mockUserSvcForTeacher{})
			err := svc.DeleteTeacher(tc.id)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUpdateTeacher
// ---------------------------------------------------------------------------

func TestTeacherService_UpdateTeacher(t *testing.T) {
	anyErr := errors.New("db error")
	email := "new@example.com"

	tests := []struct {
		name        string
		id          int64
		input       *model.UpdateTeacher
		teacherRepo *mockTeacherRepo
		subjectRepo *mockSubjectRepo
		userRepo    *mockTeacherUserRepo
		wantErr     error
	}{
		{
			name:  "success",
			id:    1,
			input: validUpdateTeacher(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return true, nil },
			},
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return true, nil }},
			userRepo:    &mockTeacherUserRepo{},
			wantErr:     nil,
		},
		{
			name:        "invalid gender",
			id:          1,
			input:       func() *model.UpdateTeacher { u := validUpdateTeacher(); u.Gender = "alien"; return u }(),
			teacherRepo: &mockTeacherRepo{},
			subjectRepo: &mockSubjectRepo{},
			userRepo:    &mockTeacherUserRepo{},
			wantErr:     ErrValidationFailed,
		},
		{
			name:        "invalid working status",
			id:          1,
			input:       func() *model.UpdateTeacher { u := validUpdateTeacher(); u.WorkingStatus = "unknown"; return u }(),
			teacherRepo: &mockTeacherRepo{},
			subjectRepo: &mockSubjectRepo{},
			userRepo:    &mockTeacherUserRepo{},
			wantErr:     ErrValidationFailed,
		},
		{
			name:  "teacher not found",
			id:    99,
			input: validUpdateTeacher(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return false, nil },
			},
			subjectRepo: &mockSubjectRepo{},
			userRepo:    &mockTeacherUserRepo{},
			wantErr:     ErrNotFound,
		},
		{
			name:  "subject not found",
			id:    1,
			input: validUpdateTeacher(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return true, nil },
			},
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return false, nil }},
			userRepo:    &mockTeacherUserRepo{},
			wantErr:     ErrNotFound,
		},
		{
			name:  "email already exists",
			id:    1,
			input: func() *model.UpdateTeacher { u := validUpdateTeacher(); u.Email = &email; return u }(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return true, nil },
			},
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return true, nil }},
			userRepo:    &mockTeacherUserRepo{emailExistsFn: func(email string) (bool, error) { return true, nil }},
			wantErr:     ErrValidationFailed,
		},
		{
			name:  "exists check db error",
			id:    1,
			input: validUpdateTeacher(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return false, anyErr },
			},
			subjectRepo: &mockSubjectRepo{},
			userRepo:    &mockTeacherUserRepo{},
			wantErr:     anyErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTeacherSvc(tc.userRepo, tc.teacherRepo, tc.subjectRepo, &mockUserSvcForTeacher{})
			err := svc.UpdateTeacher(tc.id, tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}
