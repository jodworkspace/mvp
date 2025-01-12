package v1

import (
	"context"
	"gitlab.com/tokpok/mvp/internal/usecase/user"
	"gitlab.com/tokpok/mvp/pkg/httpclient"
	"gitlab.com/tokpok/mvp/pkg/logger"
	"net/http"
)

type UserHandler struct {
	u  useruc.UserUsecase
	zl *logger.ZapLogger
}

func NewUserHandler(u useruc.UserUsecase, zl *logger.ZapLogger) *UserHandler {
	return &UserHandler{u, zl}
}

func (h *UserHandler) Userinfo(w http.ResponseWriter, r *http.Request) {
	u, err := h.u.GetUserByEmail(context.TODO())
	if err != nil {
		// return error response
	}
	_ = httpclient.WriteJSON(w, r, http.StatusOK, u)
}

func (h *UserHandler) GoogleRedirectEndpoint(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) GoogleAuthorize(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {}
