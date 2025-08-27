package v1

import (
	"errors"
	"net/http"
	"strings"
	"time"

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

func (h *OAuthHandler) ExchangeToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Provider          string `json:"provider" validate:"required"`
		AuthorizationCode string `json:"authorizationCode" validate:"required"`
		CodeVerifier      string `json:"codeVerifier" validate:"required"`
		RedirectURI       string `json:"redirectUri" validate:"required"`
	}

	err := BindWithValidation(r, &input)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Details: httpx.JSON{"error": err.Error()},
		})
		return
	}

	provider := strings.ToLower(input.Provider)

	link, err := h.oauthMng.ExchangeToken(provider, input.AuthorizationCode, input.CodeVerifier, input.RedirectURI)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	user, err := h.oauthMng.GetUserInfo(provider, link)
	if err != nil {
		h.logger.Error("OAuthHandler - ExchangeToken - GetUserInfo", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	existedUser, err := h.userUC.GetByEmail(r.Context(), user.Email)
	if err != nil && !errors.Is(err, errorx.ErrUserNotFound) {
		h.logger.Error("OAuthHandler - ExchangeToken - GetByEmail", zap.String("email", user.Email), zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// onboard new user
	if existedUser == nil {
		user.ID = uuid.NewString()
		user.Active = true
		user.CreatedAt = time.Now()
		user.UpdatedAt = user.CreatedAt

		// TODO: encrypt tokens
		link.UserID = user.ID
		link.Issuer = provider
		link.CreatedAt = user.CreatedAt
		link.UpdatedAt = user.CreatedAt

		err = h.userUC.CreateUserWithLink(r.Context(), user, link)
		if err != nil {
			h.logger.Error("OAuthHandler - ExchangeToken - CreateUserWithLink", zap.Error(err))
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}

		existedUser = user
	}

	session, err := h.sessionStore.New(r, domain.SessionCookieName)
	if err != nil {
		h.logger.Error("OAuthHandler - ExchangeToken - NewSession", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get session",
			Details: httpx.JSON{"error": err.Error()},
		})
		return
	}

	session.Values[domain.SessionKeyUserID] = existedUser.ID
	err = session.Save(r, w)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to save session",
			Details: httpx.JSON{
				"error": err.Error(),
			},
		})
		return
	}

	_ = httpx.SuccessJSON(w, http.StatusOK, httpx.JSON{
		"user": existedUser,
		"link": link,
	})
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")[7:]
	_ = httpx.SuccessJSON(w, http.StatusOK, httpx.JSON{
		"accessToken": accessToken,
	})
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {}
