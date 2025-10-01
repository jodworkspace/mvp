package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gitlab.com/jodworkspace/mvp/internal/domain"
)

type Repository interface {
	Exists(ctx context.Context, col string, val any) (bool, error)
	Insert(context.Context, *domain.User, ...pgx.Tx) error
	Get(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type LinkRepository interface {
	Insert(context.Context, *domain.Link, ...pgx.Tx) error
	Get(ctx context.Context, userID, issuer string) (*domain.Link, error)
	Update(ctx context.Context, link *domain.Link) error
}
