package services

import (
	"context"
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/auth"
)

type AuthService struct {
	userRepo ports.UserRepository
}

func NewAuthService(repo ports.UserRepository) ports.AuthService {
	return &AuthService{userRepo: repo}
}

func (s *AuthService) Register(ctx context.Context, req ports.RegisterRequest) error {
	// 1. Cek apakah username sudah ada
	exist, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return err
	}
	if exist != nil {
		return errors.New("username sudah terdaftar")
	}

	// 2. Hash Password
	hash, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	// 3. Simpan User
	newUser := &domain.User{
		Name:     req.Name,
		Username: req.Username,
		Password: hash,
		Role:     "kader", // Default role
	}

	return s.userRepo.Save(ctx, newUser)
}

func (s *AuthService) Login(ctx context.Context, req ports.LoginRequest) (string, error) {
	// 1. Cari User
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("username atau password salah")
	}

	// 2. Verifikasi Password
	match, err := argon2id.ComparePasswordAndHash(req.Password, user.Password)
	if err != nil {
		return "", err
	}
	if !match {
		return "", errors.New("username atau password salah")
	}

	// 3. Generate Token
	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}
