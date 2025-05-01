package v1

import (
	"context"
	"gitlab.com/gookie/mvp/config"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/httpclient"
	"gitlab.com/gookie/mvp/pkg/httpx"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils"
	"go.uber.org/zap"
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
	zl *logger.ZapLogger,
) *OAuthHandler {
	return &OAuthHandler{
		userUC:     userUC,
		httpClient: httpClient,
		cfg:        cfg,
		logger:     zl,
	}
}

type TokenRequest struct {
	AuthorizationCode string `json:"authorizationCode" validate:"required"`
	CodeVerifier      string `json:"codeVerifier" validate:"required"`
	RedirectURI       string `json:"redirectUri" validate:"required"`
}

func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "https://github.com")
	w.WriteHeader(http.StatusFound)
}

func (h *OAuthHandler) ExchangeToken(w http.ResponseWriter, r *http.Request) {
	input := r.Context().Value("input").(*TokenRequest)

	tokenURL, err := httpx.BuildURL(h.cfg.GoogleOAuth.TokenEndpoint, map[string]string{
		"client_id":     h.cfg.GoogleOAuth.ClientID,
		"client_secret": h.cfg.GoogleOAuth.ClientSecret,
		"code":          input.AuthorizationCode,
		"code_verifier": input.CodeVerifier,
		"grant_type":    "authorization_code",
		"redirect_uri":  input.RedirectURI,
	})
	if err != nil {
		h.logger.Error("OAuthHandler - httpx.BuildURL", zap.Error(err))
		return
	}

	_, err = h.httpClient.SendGetRequest(tokenURL)
	if err != nil {
		return
	}

	_, _ = httpx.WriteJSON(w, http.StatusOK, tokenURL)
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	claims, err := utils.VerifyJWT(accessToken, h.cfg.JWT.Secret)
	if err != nil {
		_, _ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
		})
		return
	}

	u, err := h.userUC.GetUserByEmail(context.TODO(), claims.Email)
	if err != nil {
		_, _ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
		return
	}
	_, _ = httpx.WriteJSON(w, http.StatusOK, u)
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
