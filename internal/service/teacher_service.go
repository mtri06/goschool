package service

import (
	"fmt"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"strings"
)

type teacherUserRepo interface {
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
}

type teacherRepo interface {
	CreateTeacher(newTeacher *model.NewTeacher) (*model.TeacherDetails, error)
	GetTeacherByID(id int) (*model.TeacherDetails, error)
	TeacherExists(id int) (bool, error)
	UpdateTeacher(teacherID int, update *model.UpdateTeacher) error
	DeleteTeacher(teacherID int) error
	ListTeachers(params model.ListTeachersParams) ([]model.TeacherDetails, int, error)
}

type teacherSubjectRepo interface {
	Exists(id int) (bool, error)
}

type TeacherService struct {
	userRepo    teacherUserRepo
	teacherRepo teacherRepo
	subjectRepo teacherSubjectRepo
}

func NewTeacherService(userRepo teacherUserRepo, teacherRepo teacherRepo, subjectRepo teacherSubjectRepo) *TeacherService {
	return &TeacherService{
		userRepo:    userRepo,
		teacherRepo: teacherRepo,
		subjectRepo: subjectRepo,
	}
}

func (s *TeacherService) CreateTeacher(newTeacher *model.NewTeacher) (*model.TeacherDetails, error) {
	if err := validateGender(newTeacher.Gender); err != nil {
		return nil, NewError(err.Error(), "invalid_gender", ErrValidationFailed)
	}

	if err := validatePassword(newTeacher.Password); err != nil {
		return nil, NewError(err.Error(), "invalid_password", ErrValidationFailed)
	}

	exists, err := s.userRepo.UsernameExists(newTeacher.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check if username exists: %w", err)
	}
	if exists {
		return nil, NewError("username already exists", "username_exists", ErrValidationFailed)
	}

	if newTeacher.Email != nil {
		*newTeacher.Email = strings.ToLower(*newTeacher.Email)
		exists, err := s.userRepo.EmailExists(*newTeacher.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return nil, NewError("email already exists", "email_exists", ErrValidationFailed)
		}
	}

	hashedPassword, err := hashPassword(newTeacher.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	newTeacher.Password = hashedPassword

	if err := validateTeacherWorkingStatus(newTeacher.WorkingStatus); err != nil {
		return nil, NewError(err.Error(), "invalid_working_status", ErrValidationFailed)
	}

	if newTeacher.SubjectID != nil {
		if exists, err := s.subjectRepo.Exists(*newTeacher.SubjectID); err != nil {
			return nil, err
		} else if !exists {
			return nil, NewError(fmt.Sprintf("subject not found: %v", *newTeacher.SubjectID), "subject_not_found", ErrNotFound)
		}
	}

	details, err := s.teacherRepo.CreateTeacher(newTeacher)
	if err != nil {
		return nil, fmt.Errorf("failed to create teacher: %w", err)
	}
	return details, nil
}

func (s *TeacherService) GetTeacherByID(teacherID int) (*model.TeacherDetails, error) {
	teacher, err := s.teacherRepo.GetTeacherByID(teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher: %w", err)
	}
	if teacher == nil {
		return nil, NewError(fmt.Sprintf("teacher not found with id: %d", teacherID), "teacher_not_found", ErrNotFound)
	}
	return teacher, nil
}

var listTeachersAllowedOrderBy = map[string]bool{
	"id":         true,
	"name":       true,
	"updated_at": true,
	"created_at": true,
}

func (s *TeacherService) ListTeachers(params model.ListTeachersParams) ([]model.TeacherDetails, int, error) {
	if params.Pagin.Page < 1 {
		params.Pagin.Page = constant.DefaultPage
	}
	if params.Pagin.PageSize < 1 {
		params.Pagin.PageSize = constant.DefaultPageSize
	}
	if params.Pagin.PageSize > 100 {
		params.Pagin.PageSize = 100
	}

	for _, order := range params.OrderBy {
		if !listTeachersAllowedOrderBy[order.Field] {
			return nil, 0, NewError(fmt.Sprintf("invalid order by field: %s", order.Field), "invalid_order_by_field", ErrValidationFailed)
		}
	}
	params.OrderBy = append(params.OrderBy, model.Order{Field: "id"})

	teachers, total, err := s.teacherRepo.ListTeachers(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list teachers: %w", err)
	}

	if teachers == nil {
		teachers = []model.TeacherDetails{}
	}

	return teachers, total, nil
}

func (s *TeacherService) UpdateTeacher(teacherID int, update *model.UpdateTeacher) error {
	if err := validateGender(update.Gender); err != nil {
		return NewError(err.Error(), "invalid_gender", ErrValidationFailed)
	}

	if err := validateTeacherWorkingStatus(update.WorkingStatus); err != nil {
		return NewError(err.Error(), "invalid_working_status", ErrValidationFailed)
	}

	exists, err := s.teacherRepo.TeacherExists(teacherID)
	if err != nil {
		return fmt.Errorf("failed to check if teacher exists: %w", err)
	}
	if !exists {
		return NewError("teacher not found", "teacher_not_found", ErrNotFound)
	}

	if update.SubjectID != nil {
		if exists, err := s.subjectRepo.Exists(*update.SubjectID); err != nil {
			return fmt.Errorf("failed to check if subject exists: %w", err)
		} else if !exists {
			return NewError("subject not found", "subject_not_found", ErrNotFound)
		}
	}

	if update.Email != nil {
		email := strings.ToLower(*update.Email)
		update.Email = &email
		exists, err := s.userRepo.EmailExists(email)
		if err != nil {
			return fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return NewError("email already exists", "email_exists", ErrValidationFailed)
		}
	}

	if err := s.teacherRepo.UpdateTeacher(teacherID, update); err != nil {
		return fmt.Errorf("failed to update teacher: %w", err)
	}
	return nil
}

func (s *TeacherService) DeleteTeacher(teacherID int) error {
	exists, err := s.teacherRepo.TeacherExists(teacherID)
	if err != nil {
		return fmt.Errorf("failed to check if teacher exists: %w", err)
	}
	if !exists {
		return NewError("teacher not found", "teacher_not_found", ErrNotFound)
	}

	if err := s.teacherRepo.DeleteTeacher(teacherID); err != nil {
		return fmt.Errorf("failed to delete teacher: %w", err)
	}
	return nil
}
