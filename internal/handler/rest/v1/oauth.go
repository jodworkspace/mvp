package v1

import (
	"github.com/google/uuid"
	authuc "gitlab.com/gookie/mvp/internal/usecase/auth"
	"gitlab.com/gookie/mvp/internal/usecase/oauth"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type OAuthHandler struct {
	userUC  *useruc.UseCase
	oauthUC *oauthuc.Manager
	authUC  *authuc.UseCase
	logger  *logger.ZapLogger
}

func NewOAuthHandler(userUC *useruc.UseCase, oauthUC *oauthuc.Manager, authUC *authuc.UseCase, zl *logger.ZapLogger) *OAuthHandler {
	return &OAuthHandler{
		userUC:  userUC,
		oauthUC: oauthUC,
		authUC:  authUC,
		logger:  zl,
	}
}

func (h *OAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Provider          string `json:"provider" validate:"required,oneof=google github"`
		AuthorizationCode string `json:"authorizationCode" validate:"required"`
		CodeVerifier      string `json:"codeVerifier" validate:"required"`
		RedirectURI       string `json:"redirectUri" validate:"required"`
	}

	err := BindWithValidation(r, &input)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request",
			Details:    httpx.JSON{"error": err.Error()},
		})
		return
	}

	// Call /token endpoint to exchange the authorization code for an access token
	tokens, err := h.oauthUC.ExchangeToken(input.Provider, input.AuthorizationCode, input.CodeVerifier, input.RedirectURI)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to exchange token",
		})
		return
	}

	user, fedUser, err := h.oauthUC.GetUserInfo(input.Provider, tokens[0])
	if err != nil {
		h.logger.Error("OAuthHandler - Login - GetUserInfo", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to get user info",
		})
		return
	}

	// TODO: Check if user exists in the database (by email)

	// Onboard the user if they don't exist
	user.ID = uuid.NewString()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	fedUser.UserID = user.ID
	fedUser.RefreshToken = tokens[1]
	fedUser.CreatedAt = user.CreatedAt
	fedUser.UpdatedAt = user.CreatedAt

	err = h.userUC.OnboardUser(r.Context(), user, fedUser)
	if err != nil {
		h.logger.Error("OAuthHandler - Login - OnboardUser", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to onboard user",
		})
		return
	}

	accessToken := h.authUC.GenerateToken(user.Email)
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
		"data": httpx.JSON{
			"user": user,
		},
	})
	if err != nil {
		h.logger.Error("OAuthHandler - Login - httpx.WriteJSON", zap.Error(err))
	}
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	_ = httpx.WriteJSON(w, http.StatusOK, accessToken)
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
