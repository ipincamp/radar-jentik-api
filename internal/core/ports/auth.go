package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindAll(ctx context.Context) ([]*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	GetPaginatedUsers(ctx context.Context, page int, limit int) ([]*domain.User, int64, error)
}

type AuthService interface {
	Register(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, username, password string) (string, string, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, id string, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
	GetPaginatedUsers(ctx context.Context, page int, limit int) ([]*domain.User, int64, error)
}
