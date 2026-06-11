package service

import (
	"fmt"
	"strings"

	"goschool/pkg/constant"
	"goschool/pkg/model"
)

type studentSvcUserRepo interface {
	EmailExists(email string, excludeIDs ...int) (bool, error)
	UsernameExists(username string) (bool, error)
}

type studentSvcStudentRepo interface {
	CreateStudent(newStudent *model.NewStudent) (*model.StudentDetails, error)
	GetStudentByID(id int) (*model.StudentDetails, error)
	StudentExists(id int) (bool, error)
	UpdateStudent(studentID int, update *model.UpdateStudent) error
	DeleteStudent(studentID int) error
	ListStudents(params model.ListStudentsParams) ([]model.StudentDetails, int, error)
}

type studentSvcClassRepo interface {
	ClassExists(id int) (bool, error)
}

type StudentService struct {
	userRepo    studentSvcUserRepo
	studentRepo studentSvcStudentRepo
	classRepo   studentSvcClassRepo
}

func NewStudentService(
	userRepo studentSvcUserRepo, studentRepo studentSvcStudentRepo, classRepo studentSvcClassRepo,
) *StudentService {
	return &StudentService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
		classRepo:   classRepo,
	}
}

func (s *StudentService) CreateStudent(newStudent *model.NewStudent) (*model.StudentDetails, error) {
	if err := validateGender(newStudent.Gender); err != nil {
		return nil, NewError(err.Error(), "invalid_gender", ErrValidationFailed)
	}

	if err := validatePassword(newStudent.Password); err != nil {
		return nil, NewError(err.Error(), "invalid_password", ErrValidationFailed)
	}

	exists, err := s.userRepo.UsernameExists(newStudent.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check if username exists: %w", err)
	}
	if exists {
		return nil, NewError("username already exists", "username_exists", ErrValidationFailed)
	}

	if newStudent.Email != nil {
		*newStudent.Email = strings.ToLower(*newStudent.Email)
		exists, err := s.userRepo.EmailExists(*newStudent.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return nil, NewError("email already exists", "email_exists", ErrValidationFailed)
		}
	}

	hashedPassword, err := hashPassword(newStudent.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	newStudent.Password = hashedPassword

	if newStudent.ClassID != nil {
		exists, err := s.classRepo.ClassExists(*newStudent.ClassID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if class exists: %w", err)
		}
		if !exists {
			return nil, NewError("class not found", "class_not_found", ErrValidationFailed)
		}
	}

	details, err := s.studentRepo.CreateStudent(newStudent)
	if err != nil {
		return nil, fmt.Errorf("failed to create student: %w", err)
	}
	return details, nil
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

var listStudentsAllowedOrderBy = map[string]bool{
	"id":         true,
	"name":       true,
	"updated_at": true,
	"created_at": true,
}

func (s *StudentService) ListStudents(params model.ListStudentsParams) ([]model.StudentDetails, int, error) {
	if params.Pagin.Page < 1 {
		params.Pagin.Page = constant.DefaultPage
	}
	if params.Pagin.PageSize < 1 || params.Pagin.PageSize > 100 {
		params.Pagin.PageSize = constant.DefaultPageSize
	}

	for _, order := range params.OrderBy {
		if !listStudentsAllowedOrderBy[order.Field] {
			return nil, 0, NewError(fmt.Sprintf("invalid order by field: %s", order.Field), "invalid_order_by_field", ErrValidationFailed)
		}
	}
	params.OrderBy = append(params.OrderBy, model.Order{Field: "id"})

	return s.studentRepo.ListStudents(params)
}

func (s *StudentService) UpdateStudent(studentID int, update *model.UpdateStudent) error {
	if err := validateGender(update.Gender); err != nil {
		return NewError(err.Error(), "invalid_gender", ErrValidationFailed)
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
