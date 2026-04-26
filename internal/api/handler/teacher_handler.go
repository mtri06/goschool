package handler

import (
	"net/http"
	"strconv"

	"goschool/pkg/constant"
	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type teacherSvc interface {
	CreateTeacher(newTeacher *model.NewTeacher) error
	GetTeacherByID(teacherID int64) (*model.TeacherDetails, error)
	UpdateTeacher(teacherID int64, update *model.UpdateTeacher) error
	ListTeachers(page, pageSize int, name, email, workingStatus string) ([]model.TeacherDetails, int, error)
	DeleteTeacher(teacherID int64) error
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
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidParam.WithMsg("invalid teacher id"))
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
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidParam.WithMsg("invalid user id"))
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
		httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidParam.WithMsg("invalid teacher id"))
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
	page := constant.DefaultPage
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidQuery.WithMsg("invalid page number"))
			return
		}
		page = p
	}
	pageSize := constant.DefaultPageSize
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidQuery.WithMsg("invalid page size"))
			return
		}
		pageSize = ps
	}
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")
	workingStatus := r.URL.Query().Get("workingStatus")

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
