package services

import (
	"errors"
	"reflect"
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
	listFn          func(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error)
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
func (m *mockTeacherRepo) ListTeachers(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize, name, email, workingStatus)
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

// ---------------------------------------------------------------------------
// TestCreateTeacher
// ---------------------------------------------------------------------------

func TestTeacherService_CreateTeacher_MustCallValidateUser(t *testing.T) {
	called := false
	userSvc := &mockUserSvcForTeacher{
		validateUserFn: func(u *model.User) error {
			called = true
			return nil
		},
	}

	subjectRepo := &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return true, nil }}

	svc := TeacherService{
		userRepo:    &mockTeacherUserRepo{},
		teacherRepo: &mockTeacherRepo{},
		subjectRepo: subjectRepo,
		userSvc:     userSvc,
	}

	err := svc.CreateTeacher(validNewTeacher())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected validateUser to be called, but it was not")
	}
}

func TestTeacherService_PasswordMustBeHashedOnCreate(t *testing.T) {
	var capturedTeacher *model.NewTeacher
	teacherRepo := &mockTeacherRepo{
		createFn: func(t *model.NewTeacher) error {
			capturedTeacher = t
			return nil
		},
	}

	subjectRepo := &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return true, nil }}

	svc := TeacherService{
		userRepo:    &mockTeacherUserRepo{},
		teacherRepo: teacherRepo,
		subjectRepo: subjectRepo,
		userSvc:     &mockUserSvcForTeacher{},
	}

	err := svc.CreateTeacher(validNewTeacher())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedTeacher == nil {
		t.Fatal("expected CreateTeacher to be not nil, but it was not")
	}
	if capturedTeacher.Password == validNewTeacher().Password {
		t.Error("expected password to be hashed, but it was not")
	}
}

func TestTeacherService_CreateTeacher(t *testing.T) {
	dbErr := errors.New("db error")

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
			name:  "valid input",
			input: validNewTeacher(),
		},
		{
			name: "working status on_leave",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.WorkingStatus = constant.WorkingStatusOnLeave
				return t
			}(),
		},
		{
			name: "working status active",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.WorkingStatus = constant.WorkingStatusActive
				return t
			}(),
		},
		{
			name: "working status inactive",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.WorkingStatus = constant.WorkingStatusInactive
				return t
			}(),
		},
		{
			name: "invalid working status",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.WorkingStatus = "abc123@"
				return t
			}(),
			wantErr: ErrValidationFailed,
		},
		{
			name:    "invalid working status",
			input:   func() *model.NewTeacher { t := validNewTeacher(); t.WorkingStatus = "unknown"; return t }(),
			wantErr: ErrValidationFailed,
		},
		{
			name: "female gender",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.Gender = constant.GenderFemale
				return t
			}(),
		},
		{
			name:        "subject not found",
			input:       validNewTeacher(),
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return false, nil }},
			wantErr:     ErrNotFound,
		},
		{
			name:        "subject repo returns db error",
			input:       validNewTeacher(),
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return false, dbErr }},
			wantErr:     dbErr,
		},
		{
			name:    "validateUser returns error",
			input:   validNewTeacher(),
			userSvc: &mockUserSvcForTeacher{validateUserFn: func(u *model.User) error { return ErrValidationFailed }},
			wantErr: ErrValidationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.subjectRepo == nil {
				// Subject exists by default
				tc.subjectRepo = &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return true, nil }}
			}
			if tc.userRepo == nil {
				tc.userRepo = &mockTeacherUserRepo{}
			}
			if tc.teacherRepo == nil {
				tc.teacherRepo = &mockTeacherRepo{}
			}
			if tc.userSvc == nil {
				tc.userSvc = &mockUserSvcForTeacher{}
			}
			svc := TeacherService{
				userRepo:    tc.userRepo,
				teacherRepo: tc.teacherRepo,
				subjectRepo: tc.subjectRepo,
				userSvc:     tc.userSvc,
			}
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
	teacher := &model.TeacherDetails{ID: 1445, Name: "John Doe"}
	dbErr := errors.New("db error")

	tests := []struct {
		name        string
		id          int64
		teacherRepo *mockTeacherRepo
		wantErr     error
		wantTeacher *model.TeacherDetails
	}{
		{
			name: "success",
			id:   1445,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int64) (*model.TeacherDetails, error) {
				return teacher, nil
			}},
			wantErr:     nil,
			wantTeacher: teacher,
		},
		{
			name: "teacher not found",
			id:   99,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int64) (*model.TeacherDetails, error) {
				return nil, nil
			}},
			wantErr:     ErrNotFound,
			wantTeacher: nil,
		},
		{
			name: "db error",
			id:   1,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int64) (*model.TeacherDetails, error) {
				return nil, dbErr
			}},
			wantErr:     dbErr,
			wantTeacher: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := TeacherService{
				userRepo:    &mockTeacherUserRepo{},
				teacherRepo: tc.teacherRepo,
				subjectRepo: &mockSubjectRepo{},
				userSvc:     &mockUserSvcForTeacher{},
			}
			got, err := svc.GetTeacherByID(tc.id)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
			if tc.wantTeacher != nil && !reflect.DeepEqual(got, tc.wantTeacher) {
				t.Errorf("expected %+v, got %+v", tc.wantTeacher, got)
			}
		})
	}
}

func TestTeacherService_GetTeacherByID_MustPassCorrectID(t *testing.T) {
	passedIDs := []int64{789, 456, 123, 34, 890}

	for _, id := range passedIDs {
		var capturedID *int64
		teacher := &model.TeacherDetails{ID: id, Name: "Jane Doe"}
		svc := TeacherService{
			userRepo: &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int64) (*model.TeacherDetails, error) {
				capturedID = &id
				return teacher, nil
			}},
			subjectRepo: &mockSubjectRepo{},
			userSvc:     &mockUserSvcForTeacher{},
		}
		_, err := svc.GetTeacherByID(id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if capturedID == nil || *capturedID != id {
			t.Fatalf("expected GetTeacherByID to be called with ID %d, got %v", id, capturedID)
		}
	}
}

// ---------------------------------------------------------------------------
// TestDeleteTeacher
// ---------------------------------------------------------------------------

func TestTeacherService_DeleteTeacher(t *testing.T) {
	dbErr := errors.New("db error")

	tests := []struct {
		name        string
		id          int64
		teacherRepo *mockTeacherRepo
		wantErr     error
	}{
		{
			name: "success",
			id:   33,
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
				teacherExistsFn: func(id int64) (bool, error) { return false, dbErr },
			},
			wantErr: dbErr,
		},
		{
			name: "delete error",
			id:   1,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return true, nil },
				deleteFn:        func(id int64) error { return dbErr },
			},
			wantErr: dbErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.teacherRepo == nil {
				tc.teacherRepo = &mockTeacherRepo{}
			}
			svc := TeacherService{
				userRepo:    &mockTeacherUserRepo{},
				teacherRepo: tc.teacherRepo,
				subjectRepo: &mockSubjectRepo{},
				userSvc:     &mockUserSvcForTeacher{},
			}
			err := svc.DeleteTeacher(tc.id)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestTeacherService_DeleteTeacher_MustPassCorrectID(t *testing.T) {
	passedIDs := []int64{11, 595, 34596, 2, 1348}

	for _, id := range passedIDs {
		var existCapturedID *int64
		var deleteCapturedID *int64

		svc := TeacherService{
			userRepo: &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) {
					existCapturedID = &id
					return true, nil
				},
				deleteFn: func(id int64) error {
					deleteCapturedID = &id
					return nil
				},
			},
			subjectRepo: &mockSubjectRepo{},
			userSvc:     &mockUserSvcForTeacher{},
		}

		err := svc.DeleteTeacher(id)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if existCapturedID == nil || *existCapturedID != id {
			t.Fatalf("expected teacherExistsFn to be called with ID %d, got %v", id, existCapturedID)
		}
		if deleteCapturedID == nil || *deleteCapturedID != id {
			t.Fatalf("expected deleteFn to be called with ID %d, got %v", id, deleteCapturedID)
		}
	}
}

// ---------------------------------------------------------------------------
// TestUpdateTeacher
// ---------------------------------------------------------------------------

func TestTeacherService_UpdateTeacher(t *testing.T) {
	dbErr := errors.New("db error")
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
		},
		{
			name: "male gender",
			id:   1,
			input: func() *model.UpdateTeacher {
				u := validUpdateTeacher()
				u.Gender = constant.GenderMale
				return u
			}(),
		},
		{
			name: "female gender",
			id:   1,
			input: func() *model.UpdateTeacher {
				u := validUpdateTeacher()
				u.Gender = constant.GenderFemale
				return u
			}(),
		},
		{
			name:    "invalid gender",
			id:      1,
			input:   func() *model.UpdateTeacher { u := validUpdateTeacher(); u.Gender = "alien"; return u }(),
			wantErr: ErrValidationFailed,
		},
		{
			name: "working status on_leave",
			id:   1,
			input: func() *model.UpdateTeacher {
				u := validUpdateTeacher()
				u.WorkingStatus = constant.WorkingStatusOnLeave
				return u
			}(),
		},
		{
			name: "working status active",
			id:   1,
			input: func() *model.UpdateTeacher {
				u := validUpdateTeacher()
				u.WorkingStatus = constant.WorkingStatusActive
				return u
			}(),
		},
		{
			name: "working status inactive",
			id:   1,
			input: func() *model.UpdateTeacher {
				u := validUpdateTeacher()
				u.WorkingStatus = constant.WorkingStatusInactive
				return u
			}(),
		},
		{
			name:    "invalid working status",
			id:      1,
			input:   func() *model.UpdateTeacher { u := validUpdateTeacher(); u.WorkingStatus = "unknown"; return u }(),
			wantErr: ErrValidationFailed,
		},
		{
			name:  "teacher not found",
			id:    99,
			input: validUpdateTeacher(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return false, nil },
			},
			wantErr: ErrNotFound,
		},
		{
			name:  "teacher exists check error",
			id:    1,
			input: validUpdateTeacher(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) { return false, dbErr },
			},
			wantErr: dbErr,
		},
		{
			name:        "subject not found",
			id:          1,
			input:       validUpdateTeacher(),
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return false, nil }},
			wantErr:     ErrNotFound,
		},
		{
			name:        "subject repo error",
			id:          1,
			input:       validUpdateTeacher(),
			subjectRepo: &mockSubjectRepo{existsFn: func(id int64) (bool, error) { return false, dbErr }},
			wantErr:     dbErr,
		},
		{
			name:  "email already exists",
			id:    1,
			input: func() *model.UpdateTeacher { u := validUpdateTeacher(); u.Email = &email; return u }(),
			userRepo: &mockTeacherUserRepo{
				emailExistsFn: func(email string) (bool, error) { return true, nil },
			},
			wantErr: ErrValidationFailed,
		},
		{
			name:  "email check error",
			id:    1,
			input: func() *model.UpdateTeacher { u := validUpdateTeacher(); u.Email = &email; return u }(),
			userRepo: &mockTeacherUserRepo{
				emailExistsFn: func(email string) (bool, error) { return false, dbErr },
			},
			wantErr: dbErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.userRepo == nil {
				tc.userRepo = &mockTeacherUserRepo{
					emailExistsFn: func(email string) (bool, error) { return false, nil },
				}
			}
			if tc.teacherRepo == nil {
				tc.teacherRepo = &mockTeacherRepo{
					teacherExistsFn: func(id int64) (bool, error) { return true, nil },
				}
			}
			if tc.subjectRepo == nil {
				tc.subjectRepo = &mockSubjectRepo{
					existsFn: func(id int64) (bool, error) { return true, nil },
				}
			}
			svc := TeacherService{
				userRepo:    tc.userRepo,
				teacherRepo: tc.teacherRepo,
				subjectRepo: tc.subjectRepo,
				userSvc:     &mockUserSvcForTeacher{},
			}
			err := svc.UpdateTeacher(tc.id, tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestTeacherService_UpdateTeacher_MustPassCorrectID(t *testing.T) {
	passedIDs := []int64{663, 345, 1334, 5777, 93843}

	for _, id := range passedIDs {
		var teacherExistsCapturedID *int64
		var updateCapturedID *int64

		svc := TeacherService{
			userRepo: &mockTeacherUserRepo{},
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int64) (bool, error) {
					teacherExistsCapturedID = &id
					return true, nil
				},
				updateFn: func(id int64, u *model.UpdateTeacher) error {
					updateCapturedID = &id
					return nil
				},
			},
			subjectRepo: &mockSubjectRepo{
				existsFn: func(id int64) (bool, error) { return true, nil },
			},
			userSvc: &mockUserSvcForTeacher{},
		}

		err := svc.UpdateTeacher(id, validUpdateTeacher())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if teacherExistsCapturedID == nil || *teacherExistsCapturedID != id {
			t.Errorf("expected teacherExistsFn to be called with ID %d, got %v", id, teacherExistsCapturedID)
		}
		if updateCapturedID == nil || *updateCapturedID != id {
			t.Errorf("expected updateFn to be called with ID %d, got %v", id, updateCapturedID)
		}
	}
}
