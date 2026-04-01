package services

import (
	"fmt"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"slices"
	"strings"
)

type TeacherSvcUserRepo interface {
	CreateTeacher(user *model.User, teacher *model.UserTeacher) error
	TeacherExists(id int64) (bool, error)
	UpdateTeacher(teacherID int64, update *model.UpdateTeacher) error
	DeleteTeacher(teacherID int64) error
	ListTeachers(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error)
	EmailExists(email string) (bool, error)
}

type TeacherSvcSubjectRepo interface {
	Exists(id int64) (bool, error)
}

type TeacherSvcUserSvc interface {
	validateUser(user *model.User) error
}

type TeacherService struct {
	userRepo    TeacherSvcUserRepo
	subjectRepo TeacherSvcSubjectRepo
	userSvc     TeacherSvcUserSvc
}

func NewTeacherService(userRepo TeacherSvcUserRepo, subjectRepo TeacherSvcSubjectRepo, userSvc TeacherSvcUserSvc) *TeacherService {
	return &TeacherService{
		userRepo:    userRepo,
		subjectRepo: subjectRepo,
		userSvc:     userSvc,
	}
}

func (s *TeacherService) CreateTeacher(newTeacher *model.NewTeacher) error {
	user := &model.User{
		Username:    newTeacher.Username,
		Password:    newTeacher.Password,
		Email:       newTeacher.Email,
		Name:        newTeacher.Name,
		Role:        constant.RoleTeacher,
		DateOfBirth: newTeacher.DateOfBirth,
		Gender:      newTeacher.Gender,
	}
	userTeacher := &model.UserTeacher{
		SubjectID:     newTeacher.SubjectID,
		HireDate:      newTeacher.HireDate,
		WorkingStatus: newTeacher.WorkingStatus,
	}

	if err := s.userSvc.validateUser(user); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	if !slices.Contains(teacherWorkingStatuses, userTeacher.WorkingStatus) {
		return fmt.Errorf("%w: working status must be one of %v, got: %s", ErrValidationFailed, teacherWorkingStatuses, userTeacher.WorkingStatus)
	}

	exists, err := s.subjectRepo.Exists(userTeacher.SubjectID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: subject not found: %d", ErrNotFound, userTeacher.SubjectID)
	}

	if err := s.userRepo.CreateTeacher(user, userTeacher); err != nil {
		return fmt.Errorf("failed to create teacher: %w", err)
	}
	return nil
}

func (s *TeacherService) ListTeachers(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userRepo.ListTeachers(page, pageSize, name, email)
}

func (s *TeacherService) UpdateTeacher(teacherID int64, update *model.UpdateTeacher) error {
	if !slices.Contains(allGenders, update.Gender) {
		return fmt.Errorf("%w: gender must be one of %v, got: %s", ErrValidationFailed, allGenders, update.Gender)
	}

	if !slices.Contains(teacherWorkingStatuses, update.WorkingStatus) {
		return fmt.Errorf("%w: working status must be one of %v, got: %s", ErrValidationFailed, teacherWorkingStatuses, update.WorkingStatus)
	}

	exists, err := s.userRepo.TeacherExists(teacherID)
	if err != nil {
		return fmt.Errorf("failed to check if teacher exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: teacher not found with id: %d", ErrNotFound, teacherID)
	}

	if exists, err := s.subjectRepo.Exists(update.SubjectID); err != nil {
		return fmt.Errorf("failed to check if subject exists: %w", err)
	} else if !exists {
		return fmt.Errorf("%w: subject not found: %d", ErrNotFound, update.SubjectID)
	}

	if update.Email != nil {
		email := strings.ToLower(*update.Email)
		update.Email = &email
		exists, err := s.userRepo.EmailExists(email)
		if err != nil {
			return fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return fmt.Errorf("%w: email already exists: %s", ErrValidationFailed, email)
		}
	}

	if err := s.userRepo.UpdateTeacher(teacherID, update); err != nil {
		return fmt.Errorf("failed to update teacher: %w", err)
	}
	return nil
}

func (s *TeacherService) DeleteTeacher(teacherID int64) error {
	if err := s.userRepo.DeleteTeacher(teacherID); err != nil {
		return fmt.Errorf("failed to delete teacher: %w", err)
	}
	return nil
}
