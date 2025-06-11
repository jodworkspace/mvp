package useruc

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"gitlab.com/gookie/mvp/internal/domain"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/cipherx"
	"gitlab.com/gookie/mvp/pkg/utils/errorx"
	"go.uber.org/zap"
)

type UseCase struct {
	userRepo  UserRepository
	linkRepo  LinkRepository
	txManager *postgresrepo.TransactionManager
	aead      *cipherx.AEAD
	logger    *logger.ZapLogger
}

func NewUserUseCase(
	userRepo UserRepository,
	federatedUserRepo LinkRepository,
	txManager *postgresrepo.TransactionManager,
	aead *cipherx.AEAD,
	logger *logger.ZapLogger,
) *UseCase {
	return &UseCase{
		userRepo:  userRepo,
		linkRepo:  federatedUserRepo,
		txManager: txManager,
		aead:      aead,
		logger:    logger,
	}
}

func (u *UseCase) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := u.userRepo.Exists(ctx, domain.ColEmail, email)
	if err != nil {
		u.logger.Error(
			"User - UseCase - ExistsByEmail - u.userRepo.Exists",
			zap.String("email", email),
			zap.Error(err),
		)
		return false, err
	}

	return exists, nil
}

func (u *UseCase) CreateUserWithLink(ctx context.Context, user *domain.User, link *domain.Link) error {
	return u.txManager.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		err := u.userRepo.Insert(ctx, user, tx)
		if err != nil {
			return err
		}

		err = u.linkRepo.Insert(ctx, link, tx)
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UseCase) CreateLink(ctx context.Context, link *domain.Link) error {
	err := u.linkRepo.Insert(ctx, link)
	if err != nil {
		u.logger.Error(
			"User - UseCase - CreateLink - u.linkRepo.Insert",
			zap.String("user_id", link.UserID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (u *UseCase) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		u.logger.Error(
			"User - UseCase - GetByEmail - u.userRepo.GetByEmail",
			zap.String("email", email),
			zap.Error(err),
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}
