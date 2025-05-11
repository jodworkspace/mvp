package oauth

import (
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/exception"
	"gitlab.com/gookie/mvp/pkg/utils/jwtx"
	"go.uber.org/zap"
)

type UseCase interface {
	Provider() string
	ExchangeToken(authorizationCode, codeVerifier, redirectURI string) (string, error)
	GetUserInfo(accessToken string) (*domain.User, error)
}

type Manager struct {
	cfg     *config.JWTConfig
	oauthUC map[string]UseCase
	logger  *logger.ZapLogger
}

func NewManager(cfg *config.JWTConfig, logger *logger.ZapLogger) *Manager {
	return &Manager{
		cfg:     cfg,
		oauthUC: make(map[string]UseCase),
		logger:  logger,
	}
}

func (m *Manager) RegisterOAuthProvider(uc UseCase) {
	if m.oauthUC == nil {
		m.oauthUC = make(map[string]UseCase)
	}
	m.oauthUC[uc.Provider()] = uc
}

func (m *Manager) GenerateToken(user *domain.User) string {
	return jwtx.GenerateToken(
		m.cfg.Expiry,
		[]byte(m.cfg.Secret),
		jwtx.WithIssuer(""),
		jwtx.WithSubject(user.ID),
	)
}

func (m *Manager) ExchangeToken(provider, authorizationCode, codeVerifier, redirectURI string) error {
	uc, exist := m.oauthUC[provider]
	if !exist {
		m.logger.Error("OAuthManager - ExchangeToken", zap.String("provider", provider))
		return exception.ErrInvalidProvider
	}

	return uc.ExchangeToken(authorizationCode, codeVerifier, redirectURI)
}

func (m *Manager) GetUserInfo(provider, accessToken string) (*domain.User, error) {
	uc, exist := m.oauthUC[provider]
	if !exist {
		return nil, exception.ErrInvalidProvider
	}

	return uc.GetUserInfo(accessToken)
}
