package useruc

import (
	"context"
	"gitlab.com/tokpok/mvp/internal/domain"
	postgresrepo "gitlab.com/tokpok/mvp/internal/repository/postgres"
	"gitlab.com/tokpok/mvp/pkg/logger"
)

type userUsecase struct {
	repo *postgresrepo.UserRepository
	zl   *logger.ZapLogger
}

func NewUserUsecase(repo *postgresrepo.UserRepository, zl *logger.ZapLogger) UserUsecase {
	return &userUsecase{repo, zl}
}

func (u *userUsecase) GetUserByEmail(ctx context.Context) (*domain.User, error) {
	user, err := u.repo.GetUserByEmail(ctx, "")
	if err != nil {
		return nil, err
	}
	return user, nil
}
