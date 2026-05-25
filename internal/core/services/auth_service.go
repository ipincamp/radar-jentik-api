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
func (s *authService) Login(ctx context.Context, username, password string) (string, string, error) {
	// 1. Cari User berdasarkan username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("username atau password salah")
	}

	// 2. Verifikasi Password
	match, err := argon2id.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return "", "", err
	}
	if !match {
		return "", "", errors.New("username atau password salah")
	}

	// 3. Generate Token dengan membawa ID dan Role
	token, err := s.tokenManager.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	return token, user.Role, nil // Return token beserta role-nya
}

// GetAllUsers untuk memenuhi kebutuhan fitur Manajemen Kader
func (s *authService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *authService) CreateUser(ctx context.Context, user *domain.User) error {
	return s.Register(ctx, user) // Menggunakan logika registrasi yang sama
}

func (s *authService) UpdateUser(ctx context.Context, id string, updatedData *domain.User) error {
	// 1. Cek apakah user ada
	existingUser, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}

	// 2. Validasi pencegahan username ganda jika diganti
	if updatedData.Username != existingUser.Username {
		exist, _ := s.userRepo.FindByUsername(ctx, updatedData.Username)
		if exist != nil {
			return errors.New("username sudah terdaftar, silakan gunakan yang lain")
		}
	}

	// 3. Masukkan data baru
	existingUser.FullName = updatedData.FullName
	existingUser.Username = updatedData.Username
	existingUser.VillageID = updatedData.VillageID

	// 4. Update password hanya jika kolom password diisi
	if updatedData.Password != "" {
		hash, err := argon2id.CreateHash(updatedData.Password, argon2id.DefaultParams)
		if err != nil {
			return err
		}
		existingUser.Password = hash
	}

	// 5. Simpan pembaruan
	return s.userRepo.Update(ctx, existingUser)
}

func (s *authService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}
