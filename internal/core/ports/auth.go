package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindAll(ctx context.Context) ([]*domain.User, error)
}

type AuthService interface {
	Register(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, username, password string) (string, string, error)
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
}
