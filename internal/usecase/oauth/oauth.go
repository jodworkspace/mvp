package oauthuc

import (
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/errorx"
	"go.uber.org/zap"
)

type UseCase interface {
	Provider() string
	ExchangeToken(authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error)
	GetUserInfo(link *domain.Link) (*domain.User, error)
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

func (m *Manager) ExchangeToken(provider, authorizationCode, codeVerifier, redirectURI string) (*domain.Link, error) {
	uc, exist := m.oauthUC[provider]
	if !exist {
		m.logger.Error("OAuthManager - Login", zap.String("provider", provider))
		return nil, errorx.ErrInvalidProvider
	}

	return uc.ExchangeToken(authorizationCode, codeVerifier, redirectURI)
}

func (m *Manager) GetUserInfo(provider string, link *domain.Link) (*domain.User, error) {
	uc, exist := m.oauthUC[provider]
	if !exist {
		return nil, errorx.ErrInvalidProvider
	}

	return uc.GetUserInfo(link)
}
