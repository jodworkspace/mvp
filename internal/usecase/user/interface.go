package useruc

import (
	"context"
	"github.com/jackc/pgx/v5"
	"gitlab.com/gookie/mvp/internal/domain"
)

type UserRepository interface {
	Insert(ctx context.Context, user *domain.User, tx ...pgx.Tx) error
}

type FederatedUserRepository interface {
	Insert(ctx context.Context, user *domain.FederatedUser, tx ...pgx.Tx) error
}
