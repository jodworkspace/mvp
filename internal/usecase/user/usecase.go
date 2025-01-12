package useruc

import (
	"context"
	"gitlab.com/gookie/mvp/internal/domain"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	"gitlab.com/gookie/mvp/pkg/logger"
)

type userUsecase struct {
	repo *postgresrepo.UserRepository
	zl   *logger.ZapLogger
}

func NewUserUsecase(repo *postgresrepo.UserRepository, zl *logger.ZapLogger) UserUsecase {
	return &userUsecase{repo, zl}
}

func (u *userUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
