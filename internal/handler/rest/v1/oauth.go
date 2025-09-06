package v1

import (
	"context"
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

	link, err := h.oauthMng.ExchangeToken(r.Context(), provider, input.AuthorizationCode, input.CodeVerifier, input.RedirectURI)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	user, err := h.oauthMng.GetUserInfo(r.Context(), provider, link)
	if err != nil {
		h.logger.Error("OAuthHandler - ExchangeToken - GetUserInfo", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	existedUser, err := h.userUC.GetUserByEmail(r.Context(), user.Email)
	if err != nil && !errors.Is(err, errorx.ErrUserNotFound) {
		h.logger.Error("OAuthHandler - ExchangeToken - GetUserByEmail", zap.String("email", user.Email), zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// TODO: encrypt tokens
	if existedUser == nil {
		existedUser, err = h.onboardUser(r.Context(), provider, user, link)
		if err != nil {
			h.logger.Error("OAuthHandler - ExchangeToken - CreateUserWithLink", zap.Error(err))
			_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
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
	session.Values[domain.SessionKeyIssuer] = provider
	session.Values[domain.SessionKeyAccessToken] = link.AccessToken
	session.Values[domain.SessionKeyRefreshToken] = link.RefreshToken

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

func (h *OAuthHandler) onboardUser(ctx context.Context, provider string, user *domain.User, link *domain.Link) (*domain.User, error) {
	user.ID = uuid.NewString()
	user.Active = true
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	link.UserID = user.ID
	link.Issuer = provider
	link.CreatedAt = user.CreatedAt
	link.UpdatedAt = user.CreatedAt

	err := h.userUC.CreateUserWithLink(ctx, user, link)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(domain.SessionKeyUserID).(string)

	user, err := h.userUC.GetUser(r.Context(), userID)
	if err != nil {
		code := http.StatusInternalServerError
		if !errors.Is(err, errorx.ErrUserNotFound) {
			code = http.StatusNotFound
		}

		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    code,
			Message: err.Error(),
		})
		return
	}

	_ = httpx.SuccessJSON(w, http.StatusOK, httpx.JSON{
		"user": user,
	})
}

func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(domain.SessionKeyUserID).(string)
	issuer := r.Context().Value(domain.SessionKeyIssuer).(string)

	err := h.userUC.UpdateLink(r.Context(), &domain.Link{
		UserID:                userID,
		Issuer:                issuer,
		AccessToken:           "",
		RefreshToken:          "",
		AccessTokenExpiredAt:  time.Now(),
		RefreshTokenExpiredAt: time.Now(),
	})

	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	session, err := h.sessionStore.Get(r, domain.SessionCookieName)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
	}

	_ = httpx.NoContent(w)
}
