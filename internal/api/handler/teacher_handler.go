package handler

import (
	"net/http"

	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/render"
)

type TeacherService interface {
	CreateTeacher(req model.NewTeacher) error
	ListTeachers(page, pageSize int, name, email string) ([]model.Teacher, int, error)
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

func (h *TeacherHandler) GetTeachers(w http.ResponseWriter, r *http.Request) {
	page, _ := httpx.GetQueryInt(r, "page")
	pageSize, _ := httpx.GetQueryInt(r, "pageSize")
	name := httpx.GetQueryStr(r, "name")
	email := httpx.GetQueryStr(r, "email")

	teachers, total, err := h.teacherSvc.ListTeachers(page, pageSize, name, email)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if teachers == nil {
		teachers = []model.Teacher{}
	}

	render.JSON(w, r, map[string]any{
		"teachers": teachers,
		"total":    total,
	})
}
