package v1

import (
	"context"
	"net/http"

	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
)

type DocumentUC interface {
	List(ctx context.Context, filter *domain.Pagination) ([]*domain.Document, error)
}

type DocumentHandler struct {
	documentUC DocumentUC
	logger     *logger.ZapLogger
}

func NewDocumentHandler(documentUC DocumentUC, logger *logger.ZapLogger) *DocumentHandler {
	return &DocumentHandler{
		documentUC: documentUC,
		logger:     logger,
	}
}

func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, _ := r.Context().Value(domain.KeyPagination).(*domain.Pagination)

	documents, err := h.documentUC.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("h.documentUC.List", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code: http.StatusInternalServerError,
		})
		return
	}

	_ = httpx.SuccessJSON(w, http.StatusOK, httpx.JSON{
		"page":      filter.Page,
		"documents": documents,
	})
}
