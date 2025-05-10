package useruc

import (
	"context"
	"gitlab.com/gookie/mvp/internal/domain"
)

type UserRepository interface {
	Insert(ctx context.Context, user *domain.User) error
}

type FederatedUserRepository interface {
	Insert(ctx context.Context, user *domain.FederatedUser) error
}
