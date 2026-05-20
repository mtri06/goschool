package handler

import (
	"net/http"

	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/render"
)

type SubjectSvc interface {
	CreateSubject(newSubject *model.NewSubject) (*model.SubjectDetails, error)
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
