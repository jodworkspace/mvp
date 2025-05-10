package useruc

import (
	"context"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/logger"
)

type UserUsecase struct {
	userRepo UserRepository
	zl       *logger.ZapLogger
}

func NewUserUsecase(repo UserRepository, zl *logger.ZapLogger) *UserUsecase {
	return &UserUsecase{repo, zl}
}

func (u *UserUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
