package useruc

import (
	"context"
	"gitlab.com/gookie/mvp/internal/domain"
)

type UserUsecase interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}
