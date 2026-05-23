package handler

import (
	"net/http"

	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type SubjectSvc interface {
	CreateSubject(newSubject *model.NewSubject) (*model.SubjectDetails, error)
	ListSubjects(status string, orderBy []string) ([]model.SubjectDetails, error)
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
	status := httpx.GetQueryOrDefault(r, "status", "")
	orderBy := httpx.GetQueryList(r, "orderBy")
	log.Debug().Str("status", status).Strs("orderBy", orderBy).Msg("Received GetAllSubjects request")

	subjects, err := h.subjectSvc.ListSubjects(status, orderBy)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	render.JSON(w, r, map[string]any{
		"subjects": subjects,
	})
}
