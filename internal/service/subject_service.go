package service

import (
	"fmt"

	"goschool/pkg/model"
)

type subjectSvcSubjectRepo interface {
	CreateSubject(newSubject *model.NewSubject) (*model.SubjectDetails, error)
	ExistsByName(name string) (bool, error)
}

type SubjectService struct {
	subjectRepo subjectSvcSubjectRepo
}

func NewSubjectService(subjectRepo subjectSvcSubjectRepo) *SubjectService {
	return &SubjectService{
		subjectRepo: subjectRepo,
	}
}

// CreateSubject creates a new subject with validation
func (s *SubjectService) CreateSubject(newSubject *model.NewSubject) (*model.SubjectDetails, error) {
	// Check if subject name already exists
	exists, err := s.subjectRepo.ExistsByName(newSubject.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if subject name exists: %w", err)
	}
	if exists {
		return nil, NewError("subject name already exists", "subject_name_exists", ErrValidationFailed)
	}

	// Create subject with status "active" by default
	subject, err := s.subjectRepo.CreateSubject(newSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}

	return subject, nil
}
