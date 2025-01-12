package v1

import (
	taskuc "gitlab.com/gookie/mvp/internal/usecase/task"
	"gitlab.com/gookie/mvp/pkg/logger"
	"net/http"
)

type TaskHandler struct {
	u  taskuc.TaskUsecase
	zl *logger.ZapLogger
}

func NewTaskHandler(u taskuc.TaskUsecase, zl *logger.ZapLogger) *TaskHandler {
	return &TaskHandler{u, zl}
}

func (h *TaskHandler) CreateNewTask(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
