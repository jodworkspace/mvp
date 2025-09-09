package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
)

type UseCase interface {
	Provider() string
	ExchangeToken(ctx context.Context, authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error)
	GetUserInfo(ctx context.Context, link *domain.Link) (*domain.User, error)
}

type GoogleUseCase struct {
	config     *config.GoogleOAuthConfig
	httpClient httpx.Client
	logger     *logger.ZapLogger
}

func NewGoogleUseCase(cfg *config.GoogleOAuthConfig, baseClient http.Client, logger *logger.ZapLogger) *GoogleUseCase {
	baseClient.Timeout = 10 * time.Second
	client := httpx.NewHTTPClient(baseClient)

	return &GoogleUseCase{
		httpClient: client,
		config:     cfg,
		logger:     logger,
	}
}

func (u *GoogleUseCase) Provider() string {
	return domain.ProviderGoogle
}

func (u *GoogleUseCase) ExchangeToken(ctx context.Context, authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error) {
	tokenURL, err := u.httpClient.BuildURLWithQuery(u.config.TokenEndpoint, map[string]string{
		"client_id":     u.config.ClientID,
		"client_secret": u.config.ClientSecret,
		"code":          authorizationCode,
		"code_verifier": codeVerifier,
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectURI,
	})

	if err != nil {
		u.logger.Error("GoogleUseCase - ExchangeToken - httpx.BuildURLWithQuery", zap.Error(err))
		return nil, err
	}

	resp, err := u.httpClient.DoRequest(ctx, "POST", tokenURL, nil)
	if err != nil {
		u.logger.Error("GoogleUseCase - ExchangeToken - httpClient.DoRequest", zap.Error(err))
		return nil, err
	}

	var respData struct {
		AccessToken           string `json:"access_token"`
		RefreshToken          string `json:"refresh_token"`
		TokenType             string `json:"token_type"`
		ExpiresIn             int    `json:"expires_in"`
		RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
		Scope                 string `json:"scope"`
		IDToken               string `json:"id_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		u.logger.Error("GoogleUseCase - ExchangeToken - json.NewDecoder.Decode", zap.Error(err))
		return nil, err
	}

	return &domain.Link{
		AccessToken:           respData.AccessToken,
		RefreshToken:          respData.RefreshToken,
		AccessTokenExpiredAt:  time.Now().Add(time.Duration(respData.ExpiresIn) * time.Second),
		RefreshTokenExpiredAt: time.Now().Add(time.Duration(respData.RefreshTokenExpiresIn) * time.Second),
	}, nil
}

func (u *GoogleUseCase) GetUserInfo(ctx context.Context, link *domain.Link) (*domain.User, error) {
	userInfoURL, err := u.httpClient.BuildURLWithQuery(u.config.UserInfoEndpoint, map[string]string{
		"access_token": link.AccessToken,
	})
	if err != nil {
		u.logger.Error("GoogleUseCase - httpx.BuildURLWithQuery", zap.Error(err))
		return nil, err
	}

	resp, err := u.httpClient.DoRequest(ctx, "GET", userInfoURL, nil)
	if err != nil {
		u.logger.Error("GoogleUseCase - httpClient.DoRequest", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var userinfo struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Picture       string `json:"picture"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userinfo)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		DisplayName:   userinfo.Name,
		Email:         userinfo.Email,
		EmailVerified: userinfo.EmailVerified,
		AvatarURL:     userinfo.Picture,
	}

	link.ExternalID = userinfo.Sub
	return user, nil
}
