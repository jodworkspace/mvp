package oauth

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
)

type GoogleUseCase struct {
	config     *config.GoogleOAuthConfig
	httpClient httpx.Client
	logger     *logger.ZapLogger
}

func NewGoogleUseCase(cfg *config.GoogleOAuthConfig, httpClient httpx.Client, logger *logger.ZapLogger) *GoogleUseCase {
	return &GoogleUseCase{
		httpClient: httpClient,
		config:     cfg,
		logger:     logger,
	}
}

func (u *GoogleUseCase) Provider() string {
	return domain.ProviderGoogle
}

func (u *GoogleUseCase) ExchangeToken(ctx context.Context, authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error) {
	tokenURL, err := httpx.BuildURL(u.config.TokenEndpoint, map[string]string{
		"client_id":     u.config.ClientID,
		"client_secret": u.config.ClientSecret,
		"code":          authorizationCode,
		"code_verifier": codeVerifier,
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectURI,
	})

	if err != nil {
		u.logger.Error("GoogleUseCase - VerifyUser - httpx.BuildURL", zap.Error(err))
		return nil, err
	}

	resp, err := u.httpClient.DoRequest(ctx, "POST", tokenURL, nil)
	if err != nil {
		u.logger.Error("GoogleUseCase - VerifyUser - httpClient.DoRequest", zap.Error(err))
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
		u.logger.Error("GoogleUseCase - VerifyUser - json.NewDecoder.Decode", zap.Error(err))
		return nil, err
	}

	now := time.Now().UTC()
	return &domain.Link{
		AccessToken:           respData.AccessToken,
		RefreshToken:          respData.RefreshToken,
		AccessTokenExpiredAt:  now.Add(time.Duration(respData.ExpiresIn) * time.Second),
		RefreshTokenExpiredAt: now.Add(time.Duration(respData.RefreshTokenExpiresIn) * time.Second),
	}, nil
}

func (u *GoogleUseCase) GetUserInfo(ctx context.Context, accessToken string) (*domain.User, string, error) {
	userInfoURL, err := httpx.BuildURL(u.config.UserInfoEndpoint, map[string]string{
		"access_token": accessToken,
	})
	if err != nil {
		u.logger.Error("GoogleUseCase - httpx.BuildURL", zap.Error(err))
		return nil, "", err
	}

	resp, err := u.httpClient.DoRequest(ctx, "GET", userInfoURL, nil)
	if err != nil {
		u.logger.Error("GoogleUseCase - httpClient.DoRequest", zap.Error(err))
		return nil, "", err
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
		return nil, "", err
	}

	user := &domain.User{
		DisplayName:   userinfo.Name,
		Email:         userinfo.Email,
		EmailVerified: userinfo.EmailVerified,
		AvatarURL:     userinfo.Picture,
	}

	return user, userinfo.Sub, nil
}
