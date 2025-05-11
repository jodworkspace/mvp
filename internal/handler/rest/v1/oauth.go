package v1

import (
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
	"net/http"
)

type UserUsecase interface{}

type OAuthUsecase interface {
	GenerateToken(user *domain.User) string
	ExchangeToken(provider, authorizationCode, codeVerifier, redirectURI string) error
	GetUserInfo(provider, accessToken string) (*domain.User, error)
}

type OAuthHandler struct {
	userUC  UserUsecase
	oauthUC OAuthUsecase
	logger  *logger.ZapLogger
}

func NewOAuthHandler(userUC UserUsecase, oauthUC OAuthUsecase, zl *logger.ZapLogger) *OAuthHandler {
	return &OAuthHandler{
		userUC:  userUC,
		oauthUC: oauthUC,
		logger:  zl,
	}
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
	err = h.oauthUC.ExchangeToken(provider, input.AuthorizationCode, input.CodeVerifier, input.RedirectURI)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to exchange token",
		})
		return
	}

	user := &domain.User{}

	accessToken := h.oauthUC.GenerateToken(user)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   false,
	})

	err = httpx.WriteJSON(w, http.StatusOK, httpx.JSON{
		"code":    http.StatusOK,
		"message": "success",
	})
	if err != nil {
		h.logger.Error("OAuthHandler - ExchangeToken - httpx.WriteJSON", zap.Error(err))
	}
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	_ = httpx.WriteJSON(w, http.StatusOK, accessToken)
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
