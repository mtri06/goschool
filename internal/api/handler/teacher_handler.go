package handler

import (
	"goschool/pkg/httpx"
	"goschool/pkg/model"
	"net/http"
)

type TeacherService interface {
	CreateTeacher(req model.NewTeacher) error
}

type TeacherHandler struct {
	teacherSvc TeacherService
	errMap     httpx.APIErrorMap
}

func NewTeacherHandler(teacherSvc TeacherService, errMap httpx.APIErrorMap) *TeacherHandler {
	return &TeacherHandler{teacherSvc: teacherSvc, errMap: errMap}
}

func (h *TeacherHandler) CreateTeacher(w http.ResponseWriter, r *http.Request) {
	newTeacher, err := httpx.DecodeBody[model.NewTeacher](r)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if err := h.teacherSvc.CreateTeacher(newTeacher); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
