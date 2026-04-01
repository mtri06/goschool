package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type TeacherSvc interface {
	CreateTeacher(newTeacher *model.NewTeacher) error
	UpdateTeacher(teacherID int64, update *model.UpdateTeacher) error
	ListTeachers(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error)
	DeleteTeacher(teacherID int64) error
}

type TeacherHandler struct {
	teacherSvc TeacherSvc
	errMap     httpx.APIErrorMap
}

func NewTeacherHandler(teacherSvc TeacherSvc, errMap httpx.APIErrorMap) *TeacherHandler {
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

func (h *TeacherHandler) DeleteTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, fmt.Errorf("%w: user id param is required", httpx.ErrInvalidQuery))
		return
	}

	if err := h.teacherSvc.DeleteTeacher(id); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TeacherHandler) UpdateTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, fmt.Errorf("%w: invalid teacher id", httpx.ErrInvalidQuery))
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
		teachers = []model.TeacherDetails{}
	}

	render.JSON(w, r, map[string]any{
		"teachers": teachers,
		"total":    total,
	})
}
