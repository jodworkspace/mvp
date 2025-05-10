package v1

import (
	"gitlab.com/gookie/mvp/config"
	taskuc "gitlab.com/gookie/mvp/internal/usecase/task"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
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

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()

	page := utils.StringToUint64(queries.Get("page"))
	if page == 0 {
		page = config.DefaultPage
	}

	pageSize := utils.StringToUint64(queries.Get("limit"))
	if pageSize == 0 {
		pageSize = config.DefaultPageSize
	}

	tasks, err := h.taskUC.List(r.Context(), page, pageSize)
	if err != nil {
		return
	}

	total, err := h.taskUC.Count(r.Context())
	if err != nil {
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
