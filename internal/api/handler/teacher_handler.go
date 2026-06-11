package handler

import (
	"net/http"

	"goschool/pkg/constant"
	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/render"
)

type TeacherSvc interface {
	CreateTeacher(newTeacher *model.NewTeacher) (*model.TeacherDetails, error)
	GetTeacherByID(teacherID int) (*model.TeacherDetails, error)
	UpdateTeacher(teacherID int, update *model.UpdateTeacher) error
	ListTeachers(params model.ListTeachersParams) ([]model.TeacherDetails, int, error)
	DeleteTeacher(teacherID int) error
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

	teacher, err := h.teacherSvc.CreateTeacher(newTeacher)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, teacher)
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
	name := httpx.GetQueryOptional(r, "name")
	email := httpx.GetQueryOptional(r, "email")
	workingStatus := httpx.GetQueryOptional(r, "workingStatus")

	order := httpx.GetQueryList(r, "order")

	params := model.ListTeachersParams{
		Filter: model.ListTeacherFilter{
			Name:          name,
			Email:         email,
			WorkingStatus: workingStatus,
		},
		Pagin:   model.NewPagination(page, pageSize),
		OrderBy: parseOrderBy(order),
	}

	teachers, total, err := h.teacherSvc.ListTeachers(params)
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
