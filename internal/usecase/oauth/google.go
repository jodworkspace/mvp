package oauthuc

import (
	"encoding/json"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"go.uber.org/zap"
	"net/http"
	"time"
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

func (u *GoogleUseCase) ExchangeToken(authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error) {
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
		AccessToken           string        `json:"access_token"`
		RefreshToken          string        `json:"refresh_token"`
		TokenType             string        `json:"token_type"`
		ExpiresIn             time.Duration `json:"expires_in"`
		RefreshTokenExpiresIn time.Duration `json:"refresh_token_expires_in"`
		Scope                 string        `json:"scope"`
		IDToken               string        `json:"id_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &domain.Link{
		AccessToken:           data.AccessToken,
		RefreshToken:          data.RefreshToken,
		AccessTokenExpiredAt:  time.Now().Add(data.ExpiresIn),
		RefreshTokenExpiredAt: time.Now().Add(data.ExpiresIn),
	}, nil
}

func (u *GoogleUseCase) GetUserInfo(link *domain.Link) (*domain.User, error) {
	userInfoURL, err := u.httpClient.BuildURL(u.config.UserInfoEndpoint, map[string]string{
		"access_token": link.AccessToken,
	})
	if err != nil {
		u.logger.Error("GoogleUseCase - httpx.BuildURL", zap.Error(err))
		return nil, err
	}

	resp, err := u.httpClient.DoRequest("GET", userInfoURL, nil)
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
