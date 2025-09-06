package oauthuc

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
	GetUserInfo(ctx context.Context, link *domain.Link) (*domain.User, error)
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

func (m *Manager) ExchangeToken(ctx context.Context, provider, authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error) {
	uc, exist := m.oauthUC[provider]
	if !exist {
		m.logger.Error("OAuthManager - ExchangeToken", zap.String("provider", provider))
		return nil, errorx.ErrInvalidProvider
	}

	return uc.ExchangeToken(ctx, authorizationCode, codeVerifier, redirectURI)
}

func (m *Manager) GetUserInfo(ctx context.Context, provider string, link *domain.Link) (*domain.User, error) {
	uc, exist := m.oauthUC[provider]
	if !exist {
		return nil, errorx.ErrInvalidProvider
	}

	return uc.GetUserInfo(ctx, link)
}
