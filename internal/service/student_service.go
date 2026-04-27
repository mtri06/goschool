package service

import (
	"fmt"
	"slices"
	"strings"

	repo "goschool/internal/repository"
	"goschool/pkg/constant"
	"goschool/pkg/model"
)

type studentSvcUserRepo interface {
	EmailExists(email string, excludeIDs ...int) (bool, error)
}

type studentSvcStudentRepo interface {
	CreateStudent(newStudent *model.NewStudent) error
	GetStudentByID(id int) (*model.StudentDetails, error)
	StudentExists(id int) (bool, error)
	UpdateStudent(studentID int, update *model.UpdateStudent) error
	DeleteStudent(studentID int) error
	ListStudents(p *repo.Pagination, userFilters repo.Filters, enrollmentFilters repo.Filters) ([]model.StudentDetails, int, error)
}

type studentSvcClassRepo interface {
	ClassExists(id int) (bool, error)
}

type studentSvcUserSvc interface {
	validateUser(user *model.User) error
}

type StudentService struct {
	userRepo    studentSvcUserRepo
	studentRepo studentSvcStudentRepo
	classRepo   studentSvcClassRepo
	userSvc     studentSvcUserSvc
}

func NewStudentService(
	userRepo studentSvcUserRepo, studentRepo studentSvcStudentRepo, classRepo studentSvcClassRepo, userSvc studentSvcUserSvc,
) *StudentService {
	return &StudentService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
		classRepo:   classRepo,
		userSvc:     userSvc,
	}
}

func (s *StudentService) CreateStudent(newStudent *model.NewStudent) error {
	user := &model.User{
		Username:    newStudent.Username,
		Password:    newStudent.Password,
		Email:       newStudent.Email,
		Name:        newStudent.Name,
		Role:        constant.RoleStudent,
		DateOfBirth: newStudent.DateOfBirth,
		Gender:      newStudent.Gender,
	}
	if err := s.userSvc.validateUser(user); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newStudent.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	newStudent.Password = hashedPassword

	if newStudent.ClassID != nil {
		exists, err := s.classRepo.ClassExists(*newStudent.ClassID)
		if err != nil {
			return fmt.Errorf("failed to check if class exists: %w", err)
		}
		if !exists {
			return NewError("class not found", "class_not_found", ErrValidationFailed)
		}
	}

	if err := s.studentRepo.CreateStudent(newStudent); err != nil {
		return fmt.Errorf("failed to create student: %w", err)
	}
	return nil
}

func (s *StudentService) GetStudentByID(studentID int) (*model.StudentDetails, error) {
	student, err := s.studentRepo.GetStudentByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student: %w", err)
	}
	if student == nil {
		return nil, NewError("student not found", "student_not_found", ErrNotFound)
	}
	return student, nil
}

func (s *StudentService) ListStudents(page, pageSize int, classID *int, graduated *bool, name, email string) ([]model.StudentDetails, int, error) {
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

	var userFilters repo.Filters
	if name != "" {
		userFilters = append(userFilters, repo.NewFilter("name", repo.OpLikeInsensitive, "%"+name+"%"))
	}
	if email != "" {
		email = strings.ToLower(email)
		userFilters = append(userFilters, repo.NewFilter("email", repo.OpLike, "%"+email+"%"))
	}

	var studentFilters repo.Filters
	if classID != nil {
		studentFilters = append(studentFilters, repo.NewFilter("class_id", repo.OpEquals, *classID))
	}
	if graduated != nil {
		studentFilters = append(studentFilters, repo.NewFilter("graduated", repo.OpEquals, *graduated))
	}

	return s.studentRepo.ListStudents(pagination, userFilters, studentFilters)
}

func (s *StudentService) UpdateStudent(studentID int, update *model.UpdateStudent) error {
	if !slices.Contains(allGenders, update.Gender) {
		return NewError(fmt.Sprintf("gender must be one of %v", allGenders), "invalid_gender", ErrValidationFailed)
	}

	exists, err := s.studentRepo.StudentExists(studentID)
	if err != nil {
		return fmt.Errorf("failed to check if student exists: %w", err)
	}
	if !exists {
		return NewError("student not found", "student_not_found", ErrNotFound)
	}

	if update.Email != nil {
		email := strings.ToLower(*update.Email)
		update.Email = &email
		exists, err := s.userRepo.EmailExists(email, studentID)
		if err != nil {
			return fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return NewError("email already exists", "email_exists", ErrValidationFailed)
		}
	}

	if err := s.studentRepo.UpdateStudent(studentID, update); err != nil {
		return fmt.Errorf("failed to update student: %w", err)
	}
	return nil
}

func (s *StudentService) DeleteStudent(studentID int) error {
	exists, err := s.studentRepo.StudentExists(studentID)
	if err != nil {
		return fmt.Errorf("failed to check if student exists: %w", err)
	}
	if !exists {
		return NewError("student not found", "student_not_found", ErrNotFound)
	}

	if err := s.studentRepo.DeleteStudent(studentID); err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}
	return nil
}
