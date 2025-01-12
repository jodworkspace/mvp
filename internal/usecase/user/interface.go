package useruc

import (
	"context"
	"gitlab.com/tokpok/mvp/internal/domain"
)

type UserUsecase interface {
	GetUserByEmail(ctx context.Context) (*domain.User, error)
}
