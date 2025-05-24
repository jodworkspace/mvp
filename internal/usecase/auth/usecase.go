package authuc

import (
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/jwtx"
)

type UseCase struct {
	cfg    *config.JWTConfig
	logger *logger.ZapLogger
}

func NewUseCase(cfg *config.JWTConfig, logger *logger.ZapLogger) *UseCase {
	return &UseCase{
		cfg:    cfg,
		logger: logger,
	}
}

func (u *UseCase) GenerateToken(sub string) string {
	accessToken := jwtx.GenerateToken(
		u.cfg.Expiry,
		[]byte(u.cfg.Secret),
		jwtx.WithIssuer(u.cfg.Issuer),
		jwtx.WithSubject(sub),
	)

	return accessToken
}
