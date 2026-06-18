package service

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"goschool/pkg/constant"
	"goschool/pkg/model"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockTeacherUserRepo struct {
	emailExistsFn    func(email string) (bool, error)
	usernameExistsFn func(username string) (bool, error)
}

func (m *mockTeacherUserRepo) EmailExists(email string) (bool, error) {
	if m.emailExistsFn != nil {
		return m.emailExistsFn(email)
	}
	return false, nil
}

func (m *mockTeacherUserRepo) UsernameExists(username string) (bool, error) {
	if m.usernameExistsFn != nil {
		return m.usernameExistsFn(username)
	}
	return false, nil
}

type mockTeacherRepo struct {
	createFn        func(t *model.NewTeacher) (*model.TeacherDetails, error)
	getByIDFn       func(id int) (*model.TeacherDetails, error)
	teacherExistsFn func(id int) (bool, error)
	updateFn        func(id int, u *model.UpdateTeacher) error
	deleteFn        func(id int) error
	listFn          func(params model.ListTeachersParams) ([]model.TeacherDetails, int, error)
}

func (m *mockTeacherRepo) CreateTeacher(t *model.NewTeacher) (*model.TeacherDetails, error) {
	if m.createFn != nil {
		return m.createFn(t)
	}
	return nil, nil
}
func (m *mockTeacherRepo) GetTeacherByID(id int) (*model.TeacherDetails, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}
func (m *mockTeacherRepo) TeacherExists(id int) (bool, error) {
	if m.teacherExistsFn != nil {
		return m.teacherExistsFn(id)
	}
	return false, nil
}
func (m *mockTeacherRepo) UpdateTeacher(id int, u *model.UpdateTeacher) error {
	if m.updateFn != nil {
		return m.updateFn(id, u)
	}
	return nil
}
func (m *mockTeacherRepo) DeleteTeacher(id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}
func (m *mockTeacherRepo) ListTeachers(params model.ListTeachersParams) ([]model.TeacherDetails, int, error) {
	if m.listFn != nil {
		return m.listFn(params)
	}
	return nil, 0, nil
}

type mockSubjectRepo struct {
	existsFn func(id int) (bool, error)
}

func (m *mockSubjectRepo) Exists(id int) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(id)
	}
	return false, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTeacherServiceWithMocks() *TeacherService {
	return &TeacherService{
		userRepo:    &mockTeacherUserRepo{},
		teacherRepo: &mockTeacherRepo{},
		subjectRepo: &mockSubjectRepo{},
	}
}

func validNewTeacher() *model.NewTeacher {
	email := "john.doe@example.com"
	subjectID := 1
	return &model.NewTeacher{
		Username:      "jdoe",
		Password:      "Password1!",
		Email:         &email,
		Name:          "John Doe",
		DateOfBirth:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        constant.GenderMale,
		SubjectID:     &subjectID,
		HireDate:      time.Now(),
		WorkingStatus: constant.WorkingStatusActive,
	}
}

func validUpdateTeacher() *model.UpdateTeacher {
	subjectID := 1
	return &model.UpdateTeacher{
		Name:          "Jane Doe",
		DateOfBirth:   time.Date(1991, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:        constant.GenderFemale,
		SubjectID:     &subjectID,
		HireDate:      time.Now(),
		WorkingStatus: constant.WorkingStatusActive,
	}
}

// ---------------------------------------------------------------------------
// TestCreateTeacher
// ---------------------------------------------------------------------------

func TestTeacherService_PasswordMustBeHashedOnCreate(t *testing.T) {
	var capturedTeacher *model.NewTeacher

	svc := newTeacherServiceWithMocks()
	svc.teacherRepo = &mockTeacherRepo{
		createFn: func(t *model.NewTeacher) (*model.TeacherDetails, error) {
			capturedTeacher = t
			return &model.TeacherDetails{}, nil
		},
	}
	svc.subjectRepo = &mockSubjectRepo{existsFn: func(id int) (bool, error) { return true, nil }}

	_, err := svc.CreateTeacher(validNewTeacher())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedTeacher == nil {
		t.Fatal("expected CreateTeacher to be not nil, but it was not")
	}
	require.NotEqual(t, validNewTeacher().Password, capturedTeacher.Password, "expected password to be hashed, but it was not")
}

func TestTeacherService_CreateTeacher(t *testing.T) {
	dbErr := errors.New("db error")

	tests := []struct {
		name        string
		input       *model.NewTeacher
		userRepo    *mockTeacherUserRepo
		teacherRepo *mockTeacherRepo
		subjectRepo *mockSubjectRepo
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
			name: "invalid gender",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.Gender = "invalid_gender"
				return t
			}(),
			wantErr: ErrValidationFailed,
		},
		{
			name: "invalid password",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.Password = "weak"
				return t
			}(),
			wantErr: ErrValidationFailed,
		},
		{
			name:  "username already exists",
			input: validNewTeacher(),
			userRepo: &mockTeacherUserRepo{
				usernameExistsFn: func(username string) (bool, error) { return true, nil },
			},
			wantErr: ErrValidationFailed,
		},
		{
			name:  "email already exists",
			input: validNewTeacher(),
			userRepo: &mockTeacherUserRepo{
				emailExistsFn: func(email string) (bool, error) { return true, nil },
			},
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
			subjectRepo: &mockSubjectRepo{existsFn: func(id int) (bool, error) { return false, nil }},
			wantErr:     ErrNotFound,
		},
		{
			name:        "subject repo returns db error",
			input:       validNewTeacher(),
			subjectRepo: &mockSubjectRepo{existsFn: func(id int) (bool, error) { return false, dbErr }},
			wantErr:     dbErr,
		},
		{
			name:  "username check fails",
			input: validNewTeacher(),
			userRepo: &mockTeacherUserRepo{
				usernameExistsFn: func(username string) (bool, error) { return false, dbErr },
			},
			wantErr: dbErr,
		},
		{
			name:  "email check fails",
			input: validNewTeacher(),
			userRepo: &mockTeacherUserRepo{
				emailExistsFn: func(email string) (bool, error) { return false, dbErr },
			},
			wantErr: dbErr,
		},
		{
			name: "nil subject id is allowed",
			input: func() *model.NewTeacher {
				t := validNewTeacher()
				t.SubjectID = nil
				return t
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.subjectRepo == nil {
				// Subject exists by default
				tc.subjectRepo = &mockSubjectRepo{existsFn: func(id int) (bool, error) { return true, nil }}
			}
			if tc.userRepo == nil {
				tc.userRepo = &mockTeacherUserRepo{}
			}
			if tc.teacherRepo == nil {
				tc.teacherRepo = &mockTeacherRepo{}
			}

			svc := newTeacherServiceWithMocks()
			svc.userRepo = tc.userRepo
			svc.teacherRepo = tc.teacherRepo
			svc.subjectRepo = tc.subjectRepo

			_, err := svc.CreateTeacher(tc.input)
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
		id          int
		teacherRepo *mockTeacherRepo
		wantErr     error
		wantTeacher *model.TeacherDetails
	}{
		{
			name: "success",
			id:   1445,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int) (*model.TeacherDetails, error) {
				return teacher, nil
			}},
			wantErr:     nil,
			wantTeacher: teacher,
		},
		{
			name: "teacher not found",
			id:   99,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int) (*model.TeacherDetails, error) {
				return nil, nil
			}},
			wantErr:     ErrNotFound,
			wantTeacher: nil,
		},
		{
			name: "db error",
			id:   1,
			teacherRepo: &mockTeacherRepo{getByIDFn: func(id int) (*model.TeacherDetails, error) {
				return nil, dbErr
			}},
			wantErr:     dbErr,
			wantTeacher: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTeacherServiceWithMocks()
			svc.teacherRepo = tc.teacherRepo

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
	passedIDs := []int{789, 456, 123, 34, 890}

	for _, id := range passedIDs {
		var capturedID *int
		teacher := &model.TeacherDetails{ID: id, Name: "Jane Doe"}

		svc := newTeacherServiceWithMocks()
		svc.teacherRepo = &mockTeacherRepo{getByIDFn: func(id int) (*model.TeacherDetails, error) {
			capturedID = &id
			return teacher, nil
		}}

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
		id          int
		teacherRepo *mockTeacherRepo
		wantErr     error
	}{
		{
			name: "success",
			id:   33,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int) (bool, error) { return true, nil },
				deleteFn:        func(id int) error { return nil },
			},
			wantErr: nil,
		},
		{
			name: "teacher not found",
			id:   99,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int) (bool, error) { return false, nil },
			},
			wantErr: ErrNotFound,
		},
		{
			name: "exists check error",
			id:   1,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int) (bool, error) { return false, dbErr },
			},
			wantErr: dbErr,
		},
		{
			name: "delete error",
			id:   1,
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int) (bool, error) { return true, nil },
				deleteFn:        func(id int) error { return dbErr },
			},
			wantErr: dbErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.teacherRepo == nil {
				tc.teacherRepo = &mockTeacherRepo{}
			}

			svc := newTeacherServiceWithMocks()
			svc.teacherRepo = tc.teacherRepo

			err := svc.DeleteTeacher(tc.id)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestTeacherService_DeleteTeacher_MustPassCorrectID(t *testing.T) {
	passedIDs := []int{11, 595, 34596, 2, 1348}

	for _, id := range passedIDs {
		var existCapturedID *int
		var deleteCapturedID *int

		svc := newTeacherServiceWithMocks()
		svc.teacherRepo = &mockTeacherRepo{
			teacherExistsFn: func(id int) (bool, error) {
				existCapturedID = &id
				return true, nil
			},
			deleteFn: func(id int) error {
				deleteCapturedID = &id
				return nil
			},
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
		id          int
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
				teacherExistsFn: func(id int) (bool, error) { return false, nil },
			},
			wantErr: ErrNotFound,
		},
		{
			name:  "teacher exists check error",
			id:    1,
			input: validUpdateTeacher(),
			teacherRepo: &mockTeacherRepo{
				teacherExistsFn: func(id int) (bool, error) { return false, dbErr },
			},
			wantErr: dbErr,
		},
		{
			name:        "subject not found",
			id:          1,
			input:       validUpdateTeacher(),
			subjectRepo: &mockSubjectRepo{existsFn: func(id int) (bool, error) { return false, nil }},
			wantErr:     ErrNotFound,
		},
		{
			name:        "subject repo error",
			id:          1,
			input:       validUpdateTeacher(),
			subjectRepo: &mockSubjectRepo{existsFn: func(id int) (bool, error) { return false, dbErr }},
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
		{
			name: "nil subject id is allowed",
			id:   1,
			input: func() *model.UpdateTeacher {
				u := validUpdateTeacher()
				u.SubjectID = nil
				return u
			}(),
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
					teacherExistsFn: func(id int) (bool, error) { return true, nil },
				}
			}
			if tc.subjectRepo == nil {
				tc.subjectRepo = &mockSubjectRepo{
					existsFn: func(id int) (bool, error) { return true, nil },
				}
			}

			svc := newTeacherServiceWithMocks()
			svc.userRepo = tc.userRepo
			svc.teacherRepo = tc.teacherRepo
			svc.subjectRepo = tc.subjectRepo

			err := svc.UpdateTeacher(tc.id, tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestTeacherService_UpdateTeacher_MustPassCorrectID(t *testing.T) {
	passedIDs := []int{663, 345, 1334, 5777, 93843}

	for _, id := range passedIDs {
		var teacherExistsCapturedID *int
		var updateCapturedID *int

		svc := newTeacherServiceWithMocks()
		svc.teacherRepo = &mockTeacherRepo{
			teacherExistsFn: func(id int) (bool, error) {
				teacherExistsCapturedID = &id
				return true, nil
			},
			updateFn: func(id int, u *model.UpdateTeacher) error {
				updateCapturedID = &id
				return nil
			},
		}
		svc.subjectRepo = &mockSubjectRepo{
			existsFn: func(id int) (bool, error) { return true, nil },
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

// ---------------------------------------------------------------------------
// TestListTeachers
// ---------------------------------------------------------------------------

func TestTeacherService_ListTeachers_RepoError(t *testing.T) {
	dbErr := errors.New("db error")

	svc := newTeacherServiceWithMocks()
	svc.teacherRepo = &mockTeacherRepo{
		listFn: func(params model.ListTeachersParams) ([]model.TeacherDetails, int, error) {
			return nil, 0, dbErr
		},
	}

	_, _, err := svc.ListTeachers(model.ListTeachersParams{Pagin: model.Pagination{Page: 10, PageSize: 11}})
	if !errors.Is(err, dbErr) {
		t.Errorf("expected error %v, got %v", dbErr, err)
	}
}

func TestTeacherService_ListTeachers_MustPassCorrectPaginationOptionToRepo(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		expect   model.Pagination
	}{
		{page: 1, pageSize: 10, expect: model.Pagination{Page: 1, PageSize: 10}},
		{page: 3, pageSize: 12, expect: model.Pagination{Page: 3, PageSize: 12}},
		{page: 15, pageSize: 34, expect: model.Pagination{Page: 15, PageSize: 34}},
		{page: 112, pageSize: 15, expect: model.Pagination{Page: 112, PageSize: 15}},
		{page: 1034, pageSize: 50, expect: model.Pagination{Page: 1034, PageSize: 50}},
		{page: 0, pageSize: 10, expect: model.Pagination{Page: constant.DefaultPage, PageSize: 10}},
		{page: -5, pageSize: 10, expect: model.Pagination{Page: constant.DefaultPage, PageSize: 10}},
		{page: 1, pageSize: 0, expect: model.Pagination{Page: 1, PageSize: constant.DefaultPageSize}},
		{page: 1, pageSize: -20, expect: model.Pagination{Page: 1, PageSize: constant.DefaultPageSize}},
		{page: 1, pageSize: 300, expect: model.Pagination{Page: 1, PageSize: 100}},
	}

	for _, tc := range tests {
		var capturedPagination *model.Pagination

		svc := newTeacherServiceWithMocks()
		svc.teacherRepo = &mockTeacherRepo{
			listFn: func(params model.ListTeachersParams) ([]model.TeacherDetails, int, error) {
				capturedPagination = &params.Pagin
				return nil, 0, nil
			},
		}

		_, _, err := svc.ListTeachers(model.ListTeachersParams{Pagin: model.Pagination{Page: tc.page, PageSize: tc.pageSize}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedPagination == nil || *capturedPagination != tc.expect {
			t.Errorf("expected pagination %+v, got %+v", tc.expect, capturedPagination)
		}
	}
}
