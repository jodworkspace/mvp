package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	postgresrepo "gitlab.com/jodworkspace/mvp/internal/repository/postgres"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
	"gitlab.com/jodworkspace/mvp/pkg/utils/errorx"
	"go.uber.org/zap"
)

type UseCase struct {
	userRepo  UserRepository
	linkRepo  LinkRepository
	txManager *postgresrepo.TransactionManager
	aead      *cipherx.AEAD
	logger    *logger.ZapLogger
}

func NewUseCase(
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

func (u *UseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := u.userRepo.Get(ctx, id)
	if err != nil {
		u.logger.Error(
			"User - UseCase - GetUser - u.userRepo.GetUser",
			zap.String("user_id", id),
			zap.Error(err),
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

func (u *UseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		u.logger.Error(
			"User - UseCase - GetUserByEmail - u.userRepo.GetUserByEmail",
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

func (u *UseCase) UpdateLink(ctx context.Context, link *domain.Link) error {
	linkDB, err := u.linkRepo.Get(ctx, link.UserID, link.Issuer)
	if err != nil {
		u.logger.Error(
			"User - UseCase - UpdateLink - u.linkRepo.Get",
			zap.String("user_id", link.UserID),
			zap.String("issuer", link.Issuer),
			zap.Error(err),
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return errorx.ErrLinkNotFound
		}

		return err
	}

	linkDB.AccessToken = link.AccessToken
	linkDB.RefreshToken = link.RefreshToken
	linkDB.AccessTokenExpiredAt = link.AccessTokenExpiredAt
	linkDB.RefreshTokenExpiredAt = link.RefreshTokenExpiredAt
	linkDB.UpdatedAt = time.Now()

	err = u.linkRepo.Update(ctx, linkDB)
	if err != nil {
		u.logger.Error(
			"User - UseCase - UpdateLink - u.linkRepo.Update",
			zap.String("user_id", link.UserID),
			zap.String("issuer", link.Issuer),
			zap.Error(err),
		)
		return err
	}

	return nil
}
