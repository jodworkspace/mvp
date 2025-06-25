package authuc

import (
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/jwtx"
)

type UseCase struct {
	cfg    *config.TokenConfig
	logger *logger.ZapLogger
}

func NewUseCase(cfg *config.TokenConfig, logger *logger.ZapLogger) *UseCase {
	return &UseCase{
		cfg:    cfg,
		logger: logger,
	}
}

func (u *UseCase) GenerateAccessToken(sub string) string {
	accessToken := jwtx.GenerateToken(
		[]byte(u.cfg.Secret),
		u.cfg.ShortExpiry,
		jwtx.WithIssuer(u.cfg.Issuer),
		jwtx.WithSubject(sub),
	)

	return accessToken
}

func (u *UseCase) GenerateRefreshToken(sub string) string {
	refreshToken := jwtx.GenerateToken(
		[]byte(u.cfg.RefreshSecret),
		u.cfg.LongExpiry,
		jwtx.WithIssuer(u.cfg.Issuer),
		jwtx.WithSubject(sub),
		jwtx.WithAudience("gookie-refresh"),
	)

	return refreshToken
}
