package v1

import (
	"gitlab.com/jodworkspace/mvp/internal/domain"
	taskuc "gitlab.com/jodworkspace/mvp/internal/usecase/task"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"net/http"
)

type TaskHandler struct {
	taskUC taskuc.TaskUseCase
	logger *logger.ZapLogger
}

func NewTaskHandler(taskUC taskuc.TaskUseCase, zl *logger.ZapLogger) *TaskHandler {
	return &TaskHandler{
		taskUC: taskUC,
		logger: zl,
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	pagination := r.Context().Value("pagination").(domain.Pagination)
	page := uint64(pagination.Page)
	pageSize := uint64(pagination.PageSize)

	tasks, err := h.taskUC.List(r.Context(), page, pageSize)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	total, err := h.taskUC.Count(r.Context())
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, httpx.JSON{
		"status": "ok",
		"page":   page,
		"total":  total,
		"tasks":  tasks,
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	w.WriteHeader(http.StatusOK)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {}
