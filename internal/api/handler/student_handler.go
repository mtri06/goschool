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

type studentSvc interface {
	CreateStudent(newStudent *model.NewStudent) error
	GetStudentByID(studentID int) (*model.StudentDetails, error)
	ListStudents(page, pageSize int, classID *int, graduated *bool, name, email string) ([]model.StudentDetails, int, error)
	UpdateStudent(studentID int, update *model.UpdateStudent) error
	DeleteStudent(studentID int) error
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
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if err := h.studentSvc.CreateStudent(newStudent); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *StudentHandler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidParam.WithMsg("invalid student id"))
		return
	}

	student, err := h.studentSvc.GetStudentByID(id)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	render.JSON(w, r, student)
}

func (h *StudentHandler) GetStudents(w http.ResponseWriter, r *http.Request) {
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

	var classID *int
	if classIDStr := r.URL.Query().Get("classId"); classIDStr != "" {
		id, err := strconv.Atoi(classIDStr)
		if err != nil {
			httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidQuery.WithMsg("invalid class id"))
			return
		}
		classID = &id
	}

	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")

	var graduated *bool
	if graduatedStr := r.URL.Query().Get("graduated"); graduatedStr != "" {
		v, err := strconv.ParseBool(graduatedStr)
		if err != nil {
			httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidQuery.WithMsg("invalid graduated value"))
			return
		}
		graduated = &v
	}

	students, total, err := h.studentSvc.ListStudents(page, pageSize, classID, graduated, name, email)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if students == nil {
		students = []model.StudentDetails{}
	}

	render.JSON(w, r, map[string]any{
		"students": students,
		"total":    total,
	})
}

func (h *StudentHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidParam.WithMsg("invalid student id"))
		return
	}

	update, err := httpx.DecodeBody[model.UpdateStudent](r)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if err := h.studentSvc.UpdateStudent(id, update); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *StudentHandler) DeleteStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidParam.WithMsg("invalid student id"))
		return
	}

	if err := h.studentSvc.DeleteStudent(id); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
