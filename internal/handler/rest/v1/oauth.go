package v1

import (
	"gitlab.com/gookie/mvp/internal/usecase/oauth"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
	"net/http"
)

type UserUsecase interface{}

type OAuthHandler struct {
	userUC  UserUsecase
	oauthUC map[string]oauth.UseCase
	logger  *logger.ZapLogger
}

func NewOAuthHandler(userUC UserUsecase, zl *logger.ZapLogger) *OAuthHandler {
	return &OAuthHandler{
		userUC:  userUC,
		oauthUC: make(map[string]oauth.UseCase),
		logger:  zl,
	}
}

func (h *OAuthHandler) RegisterOAuthProvider(uc oauth.UseCase) {
	if h.oauthUC == nil {
		h.oauthUC = make(map[string]oauth.UseCase)
	}
	h.oauthUC[uc.Provider()] = uc
}

func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "https://github.com")
	w.WriteHeader(http.StatusFound)
}

func (h *OAuthHandler) ExchangeToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AuthorizationCode string `json:"authorizationCode" validate:"required"`
		CodeVerifier      string `json:"codeVerifier" validate:"required"`
		RedirectURI       string `json:"redirectUri" validate:"required"`
	}

	err := Bind(r, &input)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request",
			Details:    httpx.JSON{"error": err.Error()},
		})
		return
	}

	provider := r.Context().Value("provider").(string)
	uc, exist := h.oauthUC[provider]
	if !exist {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid provider",
			Details:    httpx.JSON{"error": "invalid provider"},
		})
		return
	}

	err = uc.ExchangeToken(input.AuthorizationCode, input.CodeVerifier, input.RedirectURI)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to exchange token",
		})
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, httpx.JSON{
		"code":    http.StatusOK,
		"message": "success",
		"data": httpx.JSON{
			"access_token": "",
		},
	})
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	_ = httpx.WriteJSON(w, http.StatusOK, accessToken)
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
