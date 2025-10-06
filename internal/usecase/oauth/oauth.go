package oauth

import (
	"context"

	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/errorx"
	"go.uber.org/zap"
)

type UseCase interface {
	Provider() string
	ExchangeToken(ctx context.Context, authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error)
	GetUserInfo(ctx context.Context, accessToken string) (*domain.User, string, error)
}

type Manager struct {
	cfg     *config.TokenConfig
	oauthUC map[string]UseCase
	logger  *logger.ZapLogger
}

func NewManager(cfg *config.TokenConfig, logger *logger.ZapLogger) *Manager {
	return &Manager{
		cfg:     cfg,
		oauthUC: make(map[string]UseCase),
		logger:  logger,
	}
}

func (m *Manager) RegisterOAuthProvider(useCases ...UseCase) {
	if m.oauthUC == nil {
		m.oauthUC = make(map[string]UseCase)
	}

	for _, uc := range useCases {
		m.oauthUC[uc.Provider()] = uc
	}
}

func (m *Manager) VerifyUser(ctx context.Context, provider, authCode, codeVerifier, redirectURI string) (*domain.Link, *domain.User, error) {
	link, err := m.exchangeToken(ctx, provider, authCode, codeVerifier, redirectURI)
	if err != nil {
		return nil, nil, err
	}

	user, externalID, err := m.getUserInfo(ctx, provider, link.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	link.ExternalID = externalID
	link.Issuer = provider

	return link, user, err
}

func (m *Manager) exchangeToken(ctx context.Context, provider, authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error) {
	uc, exist := m.oauthUC[provider]
	if !exist {
		m.logger.Error("OAuthManager - VerifyUser", zap.String("provider", provider))
		return nil, errorx.ErrInvalidProvider
	}

	return uc.ExchangeToken(ctx, authorizationCode, codeVerifier, redirectURI)
}

func (m *Manager) getUserInfo(ctx context.Context, provider, accessToken string) (*domain.User, string, error) {
	uc, exist := m.oauthUC[provider]
	if !exist {
		return nil, "", errorx.ErrInvalidProvider
	}

	return uc.GetUserInfo(ctx, accessToken)
}
