package v1

import (
	taskuc "gitlab.com/gookie/mvp/internal/usecase/task"
	"gitlab.com/gookie/mvp/pkg/httpx"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils"
	"net/http"
)

type TaskHandler struct {
	taskUC taskuc.TaskUsecase
	logger *logger.ZapLogger
}

func NewTaskHandler(taskUC taskuc.TaskUsecase, zl *logger.ZapLogger) *TaskHandler {
	return &TaskHandler{
		taskUC: taskUC,
		logger: zl,
	}
}

const (
	defaultPage     = 1
	defaultPageSize = 10
)

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()

	page := utils.StringToUint64(queries.Get("page"))
	if page == 0 {
		page = defaultPage
	}

	pageSize := utils.StringToUint64(queries.Get("limit"))
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	tasks, err := h.taskUC.List(r.Context(), page, pageSize)
	if err != nil {
		return
	}

	total, err := h.taskUC.Count(r.Context())
	if err != nil {
		return
	}

	_, _ = httpx.WriteJSON(w, http.StatusOK, httpx.JSON{
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
