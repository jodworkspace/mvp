package oauthuc

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
	client := httpx.NewHTTPClient(http.Client{})

	return &GoogleUseCase{
		httpClient: client,
		config:     config,
		logger:     logger,
	}
}

func (u *GoogleUseCase) Provider() string {
	return domain.ProviderGoogle
}

func (u *GoogleUseCase) ExchangeToken(authorizationCode, codeVerifier, redirectURI string) ([]string, error) {
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
		return nil, err
	}

	resp, err := u.httpClient.DoRequest("POST", tokenURL, nil)
	if err != nil {
		u.logger.Error("GoogleUseCase - httpClient.DoRequest", zap.Error(err))
		return nil, err
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
		return nil, err
	}

	return []string{data.AccessToken, data.RefreshToken, data.IDToken}, nil
}

func (u *GoogleUseCase) GetUserInfo(accessToken string) (*domain.User, *domain.FederatedUser, error) {
	userInfoURL, err := u.httpClient.BuildURL(u.config.UserInfoEndpoint, map[string]string{
		"access_token": accessToken,
	})
	if err != nil {
		u.logger.Error("GoogleUseCase - httpx.BuildURL", zap.Error(err))
		return nil, nil, err
	}

	resp, err := u.httpClient.DoRequest("GET", userInfoURL, nil)
	if err != nil {
		u.logger.Error("GoogleUseCase - httpClient.DoRequest", zap.Error(err))
		return nil, nil, err
	}

	var userinfo struct {
		Issuer        string `json:"iss"`
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Picture       string `json:"picture"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userinfo)
	if err != nil {
		return nil, nil, err
	}

	user := &domain.User{
		DisplayName:   userinfo.Name,
		Email:         userinfo.Email,
		EmailVerified: userinfo.EmailVerified,
		AvatarURL:     userinfo.Picture,
	}

	federatedUser := &domain.FederatedUser{
		Issuer:      userinfo.Issuer,
		ExternalID:  userinfo.Sub,
		AccessToken: accessToken,
	}

	return user, federatedUser, nil
}
