package v1

import (
	"errors"
	"github.com/google/uuid"
	"gitlab.com/gookie/mvp/config"
	authuc "gitlab.com/gookie/mvp/internal/usecase/auth"
	"gitlab.com/gookie/mvp/internal/usecase/oauth"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/errorx"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type OAuthHandler struct {
	cfg      *config.TokenConfig
	userUC   *useruc.UseCase
	oauthMng *oauthuc.Manager
	authUC   *authuc.UseCase
	logger   *logger.ZapLogger
}

func NewOAuthHandler(userUC *useruc.UseCase, oauthMng *oauthuc.Manager, authUC *authuc.UseCase, zl *logger.ZapLogger) *OAuthHandler {
	return &OAuthHandler{
		userUC:   userUC,
		oauthMng: oauthMng,
		authUC:   authUC,
		logger:   zl,
	}
}

func (h *OAuthHandler) Login(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			AuthorizationCode string `json:"authorizationCode" validate:"required"`
			CodeVerifier      string `json:"codeVerifier" validate:"required"`
			RedirectURI       string `json:"redirectUri" validate:"required"`
		}

		err := BindWithValidation(r, &input)
		if err != nil {
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				StatusCode: http.StatusBadRequest,
				Message:    err.Error(),
				Details:    httpx.JSON{"error": err.Error()},
			})
			return
		}

		link, err := h.oauthMng.ExchangeToken(provider, input.AuthorizationCode, input.CodeVerifier, input.RedirectURI)
		if err != nil {
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			})
			return
		}

		user, err := h.oauthMng.GetUserInfo(provider, link)
		if err != nil {
			h.logger.Error("OAuthHandler - Login - GetUserInfo", zap.Error(err))
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			})
			return
		}

		userDB, err := h.userUC.GetByEmail(r.Context(), user.Email)
		if err != nil && !errors.Is(err, errorx.ErrUserNotFound) {
			h.logger.Error("OAuthHandler - Login - GetByEmail", zap.String("email", user.Email), zap.Error(err))
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			})
			return
		}

		if userDB == nil {
			user.ID = uuid.NewString()
			user.Active = true
			user.CreatedAt = time.Now()
			user.UpdatedAt = user.CreatedAt

			link.UserID = user.ID
			link.Issuer = provider
			link.CreatedAt = user.CreatedAt
			link.UpdatedAt = user.CreatedAt

			err = h.userUC.CreateUserWithLink(r.Context(), user, link)
			if err != nil {
				h.logger.Error("OAuthHandler - Login - CreateUserWithLink", zap.Error(err))
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					StatusCode: http.StatusInternalServerError,
					Message:    err.Error(),
				})
				return
			}
		}

		_ = httpx.WriteJSON(w, http.StatusOK, httpx.JSON{
			"code":    http.StatusOK,
			"message": "success",
			"data":    httpx.JSON{},
		})
	}
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	_ = httpx.WriteJSON(w, http.StatusOK, accessToken)
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
