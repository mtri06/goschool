package service

import (
	"fmt"

	"goschool/pkg/model"
)

type subjectRepo interface {
	Create(newSubject *model.NewSubject) (*model.SubjectDetails, error)
	ExistsByName(name string) (bool, error)
	GetAllSubjects(params model.GetAllSubjectsParams) ([]model.SubjectDetails, error)
}

type SubjectService struct {
	subjectRepo subjectRepo
}

func NewSubjectService(subjectRepo subjectRepo) *SubjectService {
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
	subject, err := s.subjectRepo.Create(newSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}

	return subject, nil
}

var subjectAllowedOrderBy = map[string]bool{
	"id":         true,
	"name":       true,
	"updated_at": true,
	"created_at": true,
}

// GetAllSubjects returns all subjects with optional filtering and ordering
func (s *SubjectService) GetAllSubjects(params model.GetAllSubjectsParams) ([]model.SubjectDetails, error) {
	for _, order := range params.OrderBy {
		if !subjectAllowedOrderBy[order.Field] {
			return nil, NewError(fmt.Sprintf("invalid order by field: %s", order.Field), "invalid_order_by_field", ErrValidationFailed)
		}
	}
	params.OrderBy = append(params.OrderBy, model.Order{Field: "id"})

	subjects, err := s.subjectRepo.GetAllSubjects(params)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}

	if subjects == nil {
		subjects = []model.SubjectDetails{}
	}

	return subjects, nil
}
