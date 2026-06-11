package handler

import (
	"net/http"

	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/render"
)

type SubjectSvc interface {
	CreateSubject(newSubject *model.NewSubject) (*model.SubjectDetails, error)
	GetAllSubjects(params model.GetAllSubjectsParams) ([]model.SubjectDetails, error)
}

type SubjectHandler struct {
	subjectSvc SubjectSvc
	errMap     httpx.APIErrorMap
}

func NewSubjectHandler(subjectSvc SubjectSvc, errMap httpx.APIErrorMap) *SubjectHandler {
	return &SubjectHandler{subjectSvc: subjectSvc, errMap: errMap}
}

// CreateSubject handles POST request to create a new subject
func (h *SubjectHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	newSubject, err := httpx.DecodeBody[model.NewSubject](r)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	subject, err := h.subjectSvc.CreateSubject(newSubject)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, subject)
}

// GetAllSubjects handles GET request to retrieve all subjects with optional filtering and ordering
func (h *SubjectHandler) GetAllSubjects(w http.ResponseWriter, r *http.Request) {
	status := httpx.GetQueryOptional(r, "status")
	order := httpx.GetQueryList(r, "order")

	filter := model.ListSubjectsFilter{
		Status: status,
	}

	params := model.GetAllSubjectsParams{
		Filter:  filter,
		OrderBy: parseOrderBy(order),
	}

	subjects, err := h.subjectSvc.GetAllSubjects(params)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}
	if subjects == nil {
		subjects = []model.SubjectDetails{}
	}

	render.JSON(w, r, map[string]any{
		"subjects": subjects,
	})
}
