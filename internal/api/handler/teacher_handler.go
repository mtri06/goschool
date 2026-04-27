package handler

import (
	"net/http"

	"goschool/pkg/constant"
	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/render"
)

type teacherSvc interface {
	CreateTeacher(newTeacher *model.NewTeacher) error
	GetTeacherByID(teacherID int) (*model.TeacherDetails, error)
	UpdateTeacher(teacherID int, update *model.UpdateTeacher) error
	ListTeachers(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error)
	DeleteTeacher(teacherID int) error
}

type TeacherHandler struct {
	teacherSvc teacherSvc
	errMap     httpx.APIErrorMap
}

func NewTeacherHandler(teacherSvc teacherSvc, errMap httpx.APIErrorMap) *TeacherHandler {
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

func (h *TeacherHandler) GetTeacherByID(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.GetParamInt(r, "id")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	teacher, err := h.teacherSvc.GetTeacherByID(id)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	render.JSON(w, r, teacher)
}

func (h *TeacherHandler) DeleteTeacher(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.GetParamInt(r, "id")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if err := h.teacherSvc.DeleteTeacher(id); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TeacherHandler) UpdateTeacher(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.GetParamInt(r, "id")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	update, err := httpx.DecodeBody[model.UpdateTeacher](r)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if err := h.teacherSvc.UpdateTeacher(id, update); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TeacherHandler) GetTeachers(w http.ResponseWriter, r *http.Request) {
	page, err := httpx.GetQueryIntOrDefault(r, "page", constant.DefaultPage)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}
	pageSize, err := httpx.GetQueryIntOrDefault(r, "pageSize", constant.DefaultPageSize)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}
	name := httpx.GetQueryOrDefault(r, "name", "")
	email := httpx.GetQueryOrDefault(r, "email", "")
	workingStatus := httpx.GetQueryOrDefault(r, "workingStatus", "")

	teachers, total, err := h.teacherSvc.ListTeachers(page, pageSize, name, email, workingStatus)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if teachers == nil {
		teachers = []model.TeacherDetails{}
	}

	render.JSON(w, r, map[string]any{
		"teachers": teachers,
		"total":    total,
	})
}
