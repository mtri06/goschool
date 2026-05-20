package handler

import (
	"net/http"

	"goschool/pkg/constant"
	"goschool/pkg/httpx"
	"goschool/pkg/model"

	"github.com/go-chi/render"
)

type StudentSvc interface {
	CreateStudent(newStudent *model.NewStudent) (*model.StudentDetails, error)
	GetStudentByID(studentID int) (*model.StudentDetails, error)
	ListStudents(page, pageSize int, classID *int, graduated *bool, name, email string) ([]model.StudentDetails, int, error)
	UpdateStudent(studentID int, update *model.UpdateStudent) error
	DeleteStudent(studentID int) error
}

type StudentHandler struct {
	studentSvc StudentSvc
	errMap     httpx.APIErrorMap
}

func NewStudentHandler(studentSvc StudentSvc, errMap httpx.APIErrorMap) *StudentHandler {
	return &StudentHandler{studentSvc: studentSvc, errMap: errMap}
}

func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	newStudent, err := httpx.DecodeBody[model.NewStudent](r)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	student, err := h.studentSvc.CreateStudent(newStudent)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, student)
}

func (h *StudentHandler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.GetParamInt(r, "id")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
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

	classID, err := httpx.GetQueryIntOptional(r, "classId")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}
	graduated, err := httpx.GetQueryBoolOptional(r, "graduated")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}
	name := httpx.GetQueryOrDefault(r, "name", "")
	email := httpx.GetQueryOrDefault(r, "email", "")

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
	id, err := httpx.GetParamInt(r, "id")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
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
	id, err := httpx.GetParamInt(r, "id")
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	if err := h.studentSvc.DeleteStudent(id); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
