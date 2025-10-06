package v1

import (
	"context"
	"net/http"

	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/helper"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
)

type TaskUC interface {
	Count(ctx context.Context, ownerID string) (int64, error)
	List(ctx context.Context, page, pageSize uint64, ownerID string) ([]*domain.Task, error)
	Create(ctx context.Context, task *domain.Task) error
	Get(ctx context.Context, id string) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id string) error
}

type TaskHandler struct {
	taskUC TaskUC
	logger *logger.ZapLogger
}

func NewTaskHandler(taskUC TaskUC, zl *logger.ZapLogger) *TaskHandler {
	return &TaskHandler{
		taskUC: taskUC,
		logger: zl,
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(domain.KeyPagination).(*domain.Pagination)
	if !ok {
		p = &domain.Pagination{
			Page:     1,
			PageSize: 10,
		}
	}

	ownerID, ok := r.Context().Value(domain.KeyUserID).(string)
	if !ok {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code: http.StatusInternalServerError,
		})
		return
	}

	tasks, err := h.taskUC.List(r.Context(), p.Page, p.PageSize, ownerID)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code: http.StatusInternalServerError,
		})
		return
	}

	total, err := h.taskUC.Count(r.Context(), ownerID)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code: http.StatusInternalServerError,
		})
		return
	}

	_ = httpx.SuccessJSON(w, http.StatusOK, httpx.JSON{
		"page":  p.Page,
		"total": total,
		"tasks": tasks,
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title     string `json:"title" validate:"required"`
		Priority  int    `json:"priority" validate:"required"`
		StartDate string `json:"startDate"`
		DueDate   string `json:"dueDate" `
	}

	if err, details := BindWithValidation(r, &input); err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Details: httpx.JSON{
				"errors": details,
			},
		})
		return
	}

	startDate, err := helper.ParseISO8601Date(input.StartDate)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid start date format",
		})
		return
	}

	dueDate, err := helper.ParseISO8601Date(input.DueDate)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid due date format",
		})
	}

	ownerID, _ := r.Context().Value("user_id").(string)
	task := &domain.Task{
		Title:     input.Title,
		Priority:  input.Priority,
		StartDate: startDate,
		DueDate:   dueDate,
		OwnerID:   ownerID,
	}

	err = h.taskUC.Create(r.Context(), task)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	_ = httpx.SuccessJSON(w, http.StatusCreated, httpx.JSON{
		"task": task,
	})
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	if taskID == "" {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "task id can not be empty",
		})
		return
	}

	err := h.taskUC.Delete(r.Context(), taskID)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	_ = httpx.WriteJSON(w, http.StatusNoContent, nil)
}
