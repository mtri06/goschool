package handler

import (
	"goschool/pkg/httpx"
	"goschool/pkg/model"
	"net/http"
)

type studentSvc interface {
	CreateStudent(newStudent *model.NewStudent) error
}

type StudentHandler struct {
	studentSvc studentSvc
	errMap     httpx.APIErrorMap
}

func NewStudentHandler(studentSvc studentSvc, errMap httpx.APIErrorMap) *StudentHandler {
	return &StudentHandler{studentSvc: studentSvc, errMap: errMap}
}

func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	newStudent, err := httpx.DecodeBody[model.NewStudent](r)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidBody.WithErr(err))
		return
	}

	if err := h.studentSvc.CreateStudent(newStudent); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
