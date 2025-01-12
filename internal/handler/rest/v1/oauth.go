package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.com/gookie/mvp/config"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/httpclient"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils"
	"io"

	"net/http"
)

type OAuthHandler struct {
	u          useruc.UserUsecase
	zl         *logger.ZapLogger
	cfg        *config.Config
	httpClient *httpclient.HTTPClient
}

func NewOAuthHandler(u useruc.UserUsecase, zl *logger.ZapLogger, cfg *config.Config) *OAuthHandler {
	httpClient := &httpclient.HTTPClient{}
	return &OAuthHandler{u, zl, cfg, httpClient}
}

func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "https://github.com")
	w.WriteHeader(http.StatusFound)
}

func (h *OAuthHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		AuthorizationCode string `json:"authorizationCode"`
	}{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.zl.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		h.zl.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	tokenURL := fmt.Sprintf("%s", h.cfg.GoogleOAuth.ClientSecret)

	_, err = h.httpClient.SendJSONRequest("GET", tokenURL, nil)
	if err != nil {
		_ = utils.ErrorJSON(w, r, nil)
	}

	_ = utils.WriteJSON(w, r, http.StatusOK, tokenURL)
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	claims, err := utils.VerifyJWT(accessToken, h.cfg.JWT.Secret)
	if err != nil {
		_ = utils.ErrorJSON(w, r, &utils.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
		})
		return
	}

	u, err := h.u.GetUserByEmail(context.TODO(), claims.Email)
	if err != nil {
		_ = utils.ErrorJSON(w, r, &utils.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
		return
	}
	_ = utils.WriteJSON(w, r, http.StatusOK, u)
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
