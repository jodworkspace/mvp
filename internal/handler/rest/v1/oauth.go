package v1

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/usecase/oauth"
	useruc "gitlab.com/jodworkspace/mvp/internal/usecase/user"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/errorx"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type OAuthHandler struct {
	cfg          *config.TokenConfig
	sessionStore sessions.Store
	userUC       *useruc.UseCase
	oauthMng     *oauthuc.Manager
	logger       *logger.ZapLogger
}

func NewOAuthHandler(
	sessionStore sessions.Store,
	userUC *useruc.UseCase,
	oauthMng *oauthuc.Manager,
	zl *logger.ZapLogger,
) *OAuthHandler {
	return &OAuthHandler{
		sessionStore: sessionStore,
		userUC:       userUC,
		oauthMng:     oauthMng,
		logger:       zl,
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

		existedUser, err := h.userUC.GetByEmail(r.Context(), user.Email)
		if err != nil && !errors.Is(err, errorx.ErrUserNotFound) {
			h.logger.Error("OAuthHandler - Login - GetByEmail", zap.String("email", user.Email), zap.Error(err))
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			})
			return
		}

		// onboard new user
		if existedUser == nil {
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

			existedUser = user
		}

		session, err := h.sessionStore.New(r, domain.SessionCookieName)
		if err != nil {
			h.logger.Error("OAuthHandler - Login - NewSession", zap.Error(err))
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to get session",
				Details:    httpx.JSON{"error": err.Error()},
			})
			return
		}

		session.Values[domain.SessionKeyUserID] = existedUser.ID
		err = session.Save(r, w)
		if err != nil {
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to save session",
				Details: httpx.JSON{
					"error": err.Error(),
				},
			})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Value:    session.ID,
			Name:     session.Name(),
			Path:     session.Options.Path,
			MaxAge:   session.Options.MaxAge,
			Expires:  time.Now().Add(time.Duration(session.Options.MaxAge) * time.Second),
			HttpOnly: session.Options.HttpOnly,
		})

		_ = httpx.SuccessJSON(w, http.StatusOK, httpx.JSON{
			"user": existedUser,
			"link": link,
		})
	}
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	_ = httpx.SuccessJSON(w, http.StatusOK, httpx.JSON{
		"accessToken": accessToken,
	})
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
