package services

import (
	"context"
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/auth"
)

type authService struct {
	userRepo     ports.UserRepository
	tokenManager *auth.TokenManager
}

func NewAuthService(repo ports.UserRepository, tm *auth.TokenManager) ports.AuthService {
	return &authService{
		userRepo:     repo,
		tokenManager: tm,
	}
}

// Register sudah diperbarui untuk langsung menerima entitas *domain.User
func (s *authService) Register(ctx context.Context, user *domain.User) error {
	// 1. Cek apakah username sudah ada
	exist, err := s.userRepo.FindByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if exist != nil {
		return errors.New("username sudah terdaftar")
	}

	// 2. Hash Password bawaan dari struct user
	hash, err := argon2id.CreateHash(user.Password, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	// Timpa password plain text dengan hash
	user.Password = hash

	// 3. Simpan User (VillageID dan FullName sudah tertanam di struct user dari handler)
	return s.userRepo.Save(ctx, user)
}

// Login sudah diperbarui untuk menerima username & password string secara langsung
func (s *authService) Login(ctx context.Context, username, password string) (string, error) {
	// 1. Cari User
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("username atau password salah")
	}

	// 2. Verifikasi Password
	match, err := argon2id.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return "", err
	}
	if !match {
		return "", errors.New("username atau password salah")
	}

	// 3. Generate Token dengan membawa ID dan Role
	token, err := s.tokenManager.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetAllUsers untuk memenuhi kebutuhan fitur Manajemen Kader
func (s *authService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.FindAll(ctx)
}
