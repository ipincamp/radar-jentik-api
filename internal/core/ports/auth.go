package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

// Driven Port (Ke Database)
type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}

// Driving Port (Ke Handler/API)
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) error
	Login(ctx context.Context, req LoginRequest) (string, error) // Return Token
}

// DTO (Data Transfer Object) untuk input
type RegisterRequest struct {
	Name     string
	Username string
	Password string
}

type LoginRequest struct {
	Username string
	Password string
}
