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

type User struct {
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
func (u *User) ToDomain() *domain.User {
	return &domain.User{
		ID:        u.ID,
		FullName:  u.Name,
		Username:  u.Username,
		Password:  u.Password,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// Helper: Mapper dari Domain ke Schema DB (Untuk Write)
func FromDomain(u *domain.User) *User {
	return &User{
		ID:       u.ID,
		Name:     u.FullName,
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
	var userDB User

	// Menggunakan struct User untuk mapping hasil query
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&userDB).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil jika user tidak ditemukan
		}
		return nil, err
	}

	// Konversi balik ke Domain agar Core tidak tahu tentang GORM
	return userDB.ToDomain(), nil
}

func (r *UserRepo) FindAll(ctx context.Context) ([]*domain.User, error) {
	var usersDB []User
	// Kecualikan password
	if err := r.db.WithContext(ctx).Omit("password").Find(&usersDB).Error; err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(usersDB))
	for i, u := range usersDB {
		users[i] = u.ToDomain()
	}
	return users, nil
}
