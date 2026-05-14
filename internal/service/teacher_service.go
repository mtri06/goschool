package service

import (
	"fmt"
	repo "goschool/internal/repository"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"strings"
)

type teacherSvcUserRepo interface {
	EmailExists(email string, excludeIDs ...int) (bool, error)
	UsernameExists(username string) (bool, error)
}

type userSvcTeacherRepo interface {
	CreateTeacher(newTeacher *model.NewTeacher) error
	GetTeacherByID(id int) (*model.TeacherDetails, error)
	TeacherExists(id int) (bool, error)
	UpdateTeacher(teacherID int, update *model.UpdateTeacher) error
	DeleteTeacher(teacherID int) error
	ListTeachers(p *repo.Pagination, userFilters repo.Filters, teacherFilters repo.Filters) ([]model.TeacherDetails, int, error)
}

type teacherSvcSubjectRepo interface {
	Exists(id int) (bool, error)
}

type TeacherService struct {
	userRepo    teacherSvcUserRepo
	teacherRepo userSvcTeacherRepo
	subjectRepo teacherSvcSubjectRepo
}

func NewTeacherService(userRepo teacherSvcUserRepo, teacherRepo userSvcTeacherRepo, subjectRepo teacherSvcSubjectRepo) *TeacherService {
	return &TeacherService{
		userRepo:    userRepo,
		teacherRepo: teacherRepo,
		subjectRepo: subjectRepo,
	}
}

func (s *TeacherService) CreateTeacher(newTeacher *model.NewTeacher) error {
	if err := validateGender(newTeacher.Gender); err != nil {
		return NewError(err.Error(), "invalid_gender", ErrValidationFailed)
	}

	if err := validatePassword(newTeacher.Password); err != nil {
		return NewError(err.Error(), "invalid_password", ErrValidationFailed)
	}

	exists, err := s.userRepo.UsernameExists(newTeacher.Username)
	if err != nil {
		return fmt.Errorf("failed to check if username exists: %w", err)
	}
	if exists {
		return NewError("username already exists", "username_exists", ErrValidationFailed)
	}

	if newTeacher.Email != nil {
		*newTeacher.Email = strings.ToLower(*newTeacher.Email)
		exists, err := s.userRepo.EmailExists(*newTeacher.Email)
		if err != nil {
			return fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return NewError("email already exists", "email_exists", ErrValidationFailed)
		}
	}

	hashedPassword, err := hashPassword(newTeacher.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	newTeacher.Password = hashedPassword

	if err := validateTeacherWorkingStatus(newTeacher.WorkingStatus); err != nil {
		return NewError(err.Error(), "invalid_working_status", ErrValidationFailed)
	}

	if newTeacher.SubjectID != nil {
		if exists, err := s.subjectRepo.Exists(*newTeacher.SubjectID); err != nil {
			return err
		} else if !exists {
			return NewError(fmt.Sprintf("subject not found: %v", *newTeacher.SubjectID), "subject_not_found", ErrNotFound)
		}
	}

	if err := s.teacherRepo.CreateTeacher(newTeacher); err != nil {
		return fmt.Errorf("failed to create teacher: %w", err)
	}
	return nil
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

func (s *TeacherService) ListTeachers(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error) {
	if page < 1 {
		page = constant.DefaultPage
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = constant.DefaultPageSize
	}
	pagination := &repo.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	userFilters := repo.Filters{
		repo.NewFilter("role", repo.OpEquals, constant.RoleTeacher),
	}
	if name != "" {
		userFilters = append(userFilters, repo.NewFilter("name", repo.OpLikeInsensitive, "%"+name+"%"))
	}
	if email != "" {
		email = strings.ToLower(email)
		userFilters = append(userFilters, repo.NewFilter("email", repo.OpLike, "%"+email+"%"))
	}
	var teacherFilters repo.Filters
	if workingStatus != "" {
		teacherFilters = append(teacherFilters, repo.NewFilter("working_status", repo.OpEquals, workingStatus))
	}

	return s.teacherRepo.ListTeachers(pagination, userFilters, teacherFilters)
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
		exists, err := s.userRepo.EmailExists(email, teacherID)
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
