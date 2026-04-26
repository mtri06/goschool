package service

import (
	"fmt"
	"goschool/pkg/constant"
	"goschool/pkg/model"
)

type studentSvcStudentRepo interface {
	CreateStudent(newStudent *model.NewStudent) error
}

type studentSvcUserSvc interface {
	validateUser(user *model.User) error
}

type StudentService struct {
	studentRepo studentSvcStudentRepo
	userSvc     studentSvcUserSvc
}

func NewStudentService(studentRepo studentSvcStudentRepo, userSvc studentSvcUserSvc) *StudentService {
	return &StudentService{studentRepo: studentRepo, userSvc: userSvc}
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

	if err := s.studentRepo.CreateStudent(newStudent); err != nil {
		return fmt.Errorf("failed to create student: %w", err)
	}
	return nil
}
