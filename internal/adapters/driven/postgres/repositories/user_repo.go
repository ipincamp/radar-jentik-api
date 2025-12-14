package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

type UserSchema struct {
	ID string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`

	Name     string `gorm:"type:varchar(100)"`
	Username string `gorm:"type:varchar(100);uniqueIndex"`
	Password string `gorm:"type:varchar(255)"`
	Role     string `gorm:"type:varchar(20)"`

	CreatedAt time.Time      `gorm:"type:timestamptz"`
	UpdatedAt time.Time      `gorm:"type:timestamptz"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
}

// Helper: Mapper dari Schema DB ke Domain (Untuk Read)
func (u *UserSchema) ToDomain() *domain.User {
	return &domain.User{
		ID:        u.ID,
		Name:      u.Name,
		Username:  u.Username,
		Password:  u.Password,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// Helper: Mapper dari Domain ke Schema DB (Untuk Write)
func FromDomain(u *domain.User) *UserSchema {
	return &UserSchema{
		ID:       u.ID,
		Name:     u.Name,
		Username: u.Username,
		Password: u.Password,
		Role:     u.Role,
	}
}

func NewUserRepo(db *gorm.DB) ports.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	// 1. Konversi Domain -> DB Schema
	userDB := FromDomain(user)

	// 2. Simpan ke Database menggunakan struct Schema
	// GORM akan melihat tag `default:uuid_generate_v4()` pada userDB.ID
	// Karena userDB.ID kosong (""), GORM tidak akan mengirim nilai '',
	// melainkan membiarkan PostgreSQL menggunakan default-nya.
	if err := r.db.WithContext(ctx).Create(userDB).Error; err != nil {
		return err
	}

	// 3. Kembalikan ID yang baru digenerate ke domain
	user.ID = userDB.ID

	return nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	// Cari berdasarkan username
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Tidak error, tapi data kosong
		}
		return nil, err
	}
	return &user, nil
}
