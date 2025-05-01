package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.com/gookie/mvp/config"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/httpclient"
	"gitlab.com/gookie/mvp/pkg/httpx"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils"
	"io"

	"net/http"
)

type OAuthHandler struct {
	userUC     useruc.UserUsecase
	httpClient *httpclient.HTTPClient
	cfg        *config.Config
	logger     *logger.ZapLogger
}

func NewOAuthHandler(
	userUC useruc.UserUsecase,
	httpClient *httpclient.HTTPClient,
	cfg *config.Config,
	zl *logger.ZapLogger) *OAuthHandler {
	return &OAuthHandler{
		userUC:     userUC,
		httpClient: httpClient,
		cfg:        cfg,
		logger:     zl,
	}
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
		h.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		h.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	tokenURL := fmt.Sprintf("%s", h.cfg.GoogleOAuth.ClientSecret)

	_, err = h.httpClient.SendJSONRequest("GET", tokenURL, nil)
	if err != nil {
		_, _ = httpx.ErrorJSON(w, nil)
	}

	_, _ = httpx.WriteJSON(w, http.StatusOK, tokenURL)
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	claims, err := utils.VerifyJWT(accessToken, h.cfg.JWT.Secret)
	if err != nil {
		_, _ = httpx.ErrorJSON(w, &httpx.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
		})
		return
	}

	u, err := h.userUC.GetUserByEmail(context.TODO(), claims.Email)
	if err != nil {
		_, _ = httpx.ErrorJSON(w, &httpx.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
		return
	}
	_, _ = httpx.WriteJSON(w, http.StatusOK, u)
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
