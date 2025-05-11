package oauth

import (
	"encoding/json"
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
	"net/http"
)

type GoogleUseCase struct {
	httpClient httpx.Client
	config     *config.GoogleOAuthConfig
	logger     *logger.ZapLogger
}

func NewGoogleUseCase(config *config.GoogleOAuthConfig, logger *logger.ZapLogger) *GoogleUseCase {
	client := httpx.NewHTTPClient(&http.Client{})

	return &GoogleUseCase{
		httpClient: client,
		config:     config,
		logger:     logger,
	}
}

func (u *GoogleUseCase) Provider() string {
	return "google"
}

func (u *GoogleUseCase) onboardUser(user *domain.User, federatedUser *domain.FederatedUser) error {
	return nil
}

func (u *GoogleUseCase) ExchangeToken(authorizationCode, codeVerifier, redirectURI string) (string, error) {
	tokenURL, err := u.httpClient.BuildURL(u.config.TokenEndpoint, map[string]string{
		"client_id":     u.config.ClientID,
		"client_secret": u.config.ClientSecret,
		"code":          authorizationCode,
		"code_verifier": codeVerifier,
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectURI,
	})

	if err != nil {
		u.logger.Error("GoogleUseCase - httpx.BuildURL", zap.Error(err))
		return "", err
	}

	resp, err := u.httpClient.DoRequest("POST", tokenURL)
	if err != nil {
		u.logger.Error("GoogleUseCase - httpClient.DoRequest", zap.Error(err))
		return "", err
	}

	var data struct {
		AccessToken           string `json:"access_token"`
		RefreshToken          string `json:"refresh_token"`
		TokenType             string `json:"token_type"`
		ExpiresIn             int    `json:"expires_in"`
		RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
		Scope                 string `json:"scope"`
		IDToken               string `json:"id_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return data.AccessToken, nil
}

func (u *GoogleUseCase) GetUserInfo(accessToken string) (*domain.User, error) {
	userInfoURL, err := u.httpClient.BuildURL(u.config.UserInfoEndpoint, map[string]string{
		"access_token": accessToken,
	})
	if err != nil {
		u.logger.Error("GoogleUseCase - httpx.BuildURL", zap.Error(err))
		return nil, err
	}

	resp, err := u.httpClient.DoRequest("GET", userInfoURL)
	if err != nil {
		u.logger.Error("GoogleUseCase - httpClient.DoRequest", zap.Error(err))
		return nil, err
	}

	var userinfo struct {
		DisplayName   string `json:"display_name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userinfo)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
