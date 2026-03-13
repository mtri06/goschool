package services

import (
	"fmt"
	"slices"

	"goschool/pkg/constant"
	"goschool/pkg/model"
)

var (
	teacherGenders         = []string{constant.GenderMale, constant.GenderFemale, constant.GenderOther}
	teacherWorkingStatuses = []string{
		constant.WorkingStatusActive,
		constant.WorkingStatusInactive,
		constant.WorkingStatusOnLeave,
	}
)

type TeacherService struct {
	teacherRepo TeacherSvcTeacherRepo
	userRepo    TeacherSvcUserRepo
	subjectRepo TeacherSvcSubjectRepo
}

type TeacherSvcTeacherRepo interface {
	Create(user *model.User, teacher *model.Teacher) error
	List(page, pageSize int, name, email string) ([]model.Teacher, int, error)
}

type TeacherSvcSubjectRepo interface {
	Exists(id int64) (bool, error)
}

type TeacherSvcUserRepo interface {
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
}

func NewTeacherService(teacherRepo TeacherSvcTeacherRepo, userRepo TeacherSvcUserRepo, subjectRepo TeacherSvcSubjectRepo) *TeacherService {
	return &TeacherService{
		teacherRepo: teacherRepo,
		userRepo:    userRepo,
		subjectRepo: subjectRepo,
	}
}

func (s *TeacherService) ListTeachers(page, pageSize int, name, email string) ([]model.Teacher, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.teacherRepo.List(page, pageSize, name, email)
}

func (s *TeacherService) CreateTeacher(req model.NewTeacher) error {
	// Business validation
	if !slices.Contains(teacherGenders, req.Gender) {
		return fmt.Errorf("%w: gender must be one of %v, got: %s", ErrValidationFailed, teacherGenders, req.Gender)
	}

	if !slices.Contains(teacherWorkingStatuses, req.WorkingStatus) {
		return fmt.Errorf("%w: working status must be one of %v, got: %s", ErrValidationFailed, teacherWorkingStatuses, req.WorkingStatus)
	}

	if err := validatePassword(req.Password); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return err
	}

	// Check if subject exists
	exists, err := s.subjectRepo.Exists(req.SubjectID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: subject not found: %d", ErrNotFound, req.SubjectID)
	}

	if req.Email != nil {
		exists, err = s.userRepo.EmailExists(*req.Email)
		if err != nil {
			return fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return fmt.Errorf("%w: email already exists: %s", ErrValidationFailed, *req.Email)
		}
	}

	exists, err = s.userRepo.UsernameExists(req.Username)
	if err != nil {
		return fmt.Errorf("failed to check if username exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: username already exists: %s", ErrValidationFailed, req.Username)
	}

	// Build user model
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Role:     constant.RoleTeacher,
	}

	// Build teacher model
	teacher := &model.Teacher{
		Name:          req.Name,
		SubjectID:     req.SubjectID,
		DateOfBirth:   req.DateOfBirth,
		Gender:        req.Gender,
		HireDate:      req.HireDate,
		WorkingStatus: req.WorkingStatus,
	}

	// Create teacher and user in a single transaction
	if err := s.teacherRepo.Create(user, teacher); err != nil {
		return fmt.Errorf("failed to create teacher: %w", err)
	}

	return nil
}
