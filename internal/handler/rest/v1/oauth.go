package v1

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/errorx"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
)

type OAuthManager interface {
	VerifyUser(ctx context.Context, provider, authCode, codeVerifier, redirectURI string) (*domain.Link, *domain.User, error)
}

type UserUC interface {
	CreateUserWithLink(ctx context.Context, user *domain.User, link *domain.Link) error
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateLink(ctx context.Context, link *domain.Link) error
}

type OAuthHandler struct {
	cfg          *config.TokenConfig
	sessionStore sessions.Store
	userUC       UserUC
	oauthMng     OAuthManager
	logger       *logger.ZapLogger
}

func NewOAuthHandler(
	sessionStore sessions.Store,
	userUC UserUC,
	oauthMng OAuthManager,
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
	var requestPayload struct {
		Provider          string `json:"provider" validate:"required"`
		AuthorizationCode string `json:"authorizationCode" validate:"required"`
		CodeVerifier      string `json:"codeVerifier" validate:"required"`
		RedirectURI       string `json:"redirectUri" validate:"required"`
	}

	err, details := BindWithValidation(r, &requestPayload)
	if err != nil {
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Details: httpx.JSON{
				"errors": details,
			},
		})
		return
	}

	provider := strings.ToLower(requestPayload.Provider)
	link, user, err := h.verifyUser(r.Context(), provider,
		requestPayload.AuthorizationCode,
		requestPayload.CodeVerifier,
		requestPayload.RedirectURI,
	)
	if err != nil {
		h.logger.Error("OAuthHandler - VerifyUser - h.onboardUser", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	session, err := h.sessionStore.New(r, domain.SessionCookieName)
	if err != nil {
		h.logger.Error("OAuthHandler - VerifyUser - h.sessionStore.New", zap.Error(err))
		_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to get session",
			Details: httpx.JSON{
				"error": err.Error(),
			},
		})
		return
	}

	session.Values[domain.KeyUserID] = user.ID
	session.Values[domain.KeyIssuer] = provider
	session.Values[domain.KeyAccessToken] = link.AccessToken
	session.Values[domain.KeyRefreshToken] = link.RefreshToken

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
		"user": user,
		"link": link,
	})
}

func (h *OAuthHandler) verifyUser(ctx context.Context, provider, authCode, codeVerifier, redirectURI string) (*domain.Link, *domain.User, error) {
	link, user, err := h.oauthMng.VerifyUser(ctx, provider, authCode, codeVerifier, redirectURI)
	if err != nil {
		return nil, nil, err
	}

	existedUser, err := h.userUC.GetUserByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, errorx.ErrUserNotFound) {
		return nil, nil, err
	}

	if existedUser != nil {
		err = h.userUC.UpdateLink(ctx, link)
		if err != nil {
			return nil, nil, err
		}

		return link, existedUser, err
	}

	err = h.userUC.CreateUserWithLink(ctx, user, link)
	if err != nil {
		return nil, nil, err
	}

	return link, user, err
}

func (h *OAuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(domain.KeyUserID).(string)

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
	userID, _ := r.Context().Value(domain.KeyUserID).(string)
	issuer, _ := r.Context().Value(domain.KeyIssuer).(string)

	now := time.Now().UTC()
	err := h.userUC.UpdateLink(r.Context(), &domain.Link{
		UserID:                userID,
		Issuer:                issuer,
		AccessToken:           "",
		RefreshToken:          "",
		AccessTokenExpiredAt:  now,
		RefreshTokenExpiredAt: now,
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
