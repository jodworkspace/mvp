package useruc

import (
	"context"
	"github.com/jackc/pgx/v5"
	"gitlab.com/gookie/mvp/internal/domain"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	"gitlab.com/gookie/mvp/pkg/logger"
	"go.uber.org/zap"
)

type UseCase struct {
	userRepo          UserRepository
	federatedUserRepo FederatedUserRepository
	txManager         *postgresrepo.TransactionManager
	logger            *logger.ZapLogger
}

func NewUserUseCase(
	userRepo UserRepository,
	federatedUserRepo FederatedUserRepository,
	txManager *postgresrepo.TransactionManager,
	logger *logger.ZapLogger,
) *UseCase {
	return &UseCase{
		userRepo:          userRepo,
		federatedUserRepo: federatedUserRepo,
		txManager:         txManager,
		logger:            logger,
	}
}

func (u *UseCase) OnboardUser(ctx context.Context, user *domain.User, federatedUser *domain.FederatedUser) error {
	return u.txManager.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		err := u.userRepo.Insert(ctx, user, tx)
		if err != nil {
			return err
		}

		federatedUser.UserID = user.ID
		err = u.federatedUserRepo.Insert(ctx, federatedUser, tx)
		if err != nil {
			return err
		}

		return nil
	})

}

func (u *UseCase) CreateFederatedLink(ctx context.Context, federatedUser *domain.FederatedUser) error {
	err := u.federatedUserRepo.Insert(ctx, federatedUser)
	if err != nil {
		u.logger.Error(
			"User - UseCase - CreateFederatedLink - u.federatedUserRepo.Insert",
			zap.String("user_id", federatedUser.UserID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (u *UseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
