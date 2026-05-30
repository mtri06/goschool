package service

import (
	"fmt"
	"strings"

	"goschool/internal/repository"
	"goschool/pkg/model"
)

type subjectSvcSubjectRepo interface {
	CreateSubject(newSubject *model.NewSubject) (*model.SubjectDetails, error)
	ExistsByName(name string) (bool, error)
	ListSubjects(status string, orderBy repository.OrderBy) ([]model.SubjectDetails, error)
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

// ListSubjects returns all subjects with optional filtering and ordering
func (s *SubjectService) ListSubjects(status string, order []string) ([]model.SubjectDetails, error) {
	var orderBy repository.OrderBy
	for _, ob := range order {
		field := strings.TrimPrefix(ob, "-")
		if field == "name" || field == "updated_at" {
			orderBy = append(orderBy, ob)
		}
	}

	subjects, err := s.subjectRepo.ListSubjects(status, orderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}

	if subjects == nil {
		subjects = []model.SubjectDetails{}
	}

	return subjects, nil
}
