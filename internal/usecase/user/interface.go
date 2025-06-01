package useruc

import (
	"context"
	"github.com/jackc/pgx/v5"
	"gitlab.com/gookie/mvp/internal/domain"
)

type UserRepository interface {
	Exists(ctx context.Context, col string, val any) (bool, error)
	Insert(context.Context, *domain.User, ...pgx.Tx) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type LinkRepository interface {
	Insert(context.Context, *domain.Link, ...pgx.Tx) error
}
