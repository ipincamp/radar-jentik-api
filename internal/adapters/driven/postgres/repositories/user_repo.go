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

// Skema DB lokal
type User struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name      string         `gorm:"type:varchar(100)"`
	Username  string         `gorm:"type:varchar(100);uniqueIndex"`
	Password  string         `gorm:"type:varchar(255)"`
	Role      string         `gorm:"type:varchar(20)"`
	VillageID string         `gorm:"type:uuid;not null"`
	CreatedAt time.Time      `gorm:"type:timestamptz"`
	UpdatedAt time.Time      `gorm:"type:timestamptz"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`

	// Tambahkan Relasi ke Village agar bisa di-Preload (Join)
	Village *domain.Village `gorm:"foreignKey:VillageID"`
}

// Mapper: Skema DB -> Domain
func (u *User) ToDomain() *domain.User {
	return &domain.User{
		ID:        u.ID,
		FullName:  u.Name,
		Username:  u.Username,
		Password:  u.Password,
		Role:      u.Role,
		VillageID: u.VillageID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Village:   u.Village, // Bawa data relasi desa
	}
}

// Mapper: Domain -> Skema DB
func FromDomain(d *domain.User) *User {
	return &User{
		ID:        d.ID,
		Name:      d.FullName,
		Username:  d.Username,
		Password:  d.Password,
		Role:      d.Role,
		VillageID: d.VillageID,
	}
}

func NewUserRepo(db *gorm.DB) ports.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	userDB := FromDomain(user)
	if err := r.db.WithContext(ctx).Create(userDB).Error; err != nil {
		return err
	}

	user.ID = userDB.ID
	user.CreatedAt = userDB.CreatedAt
	user.UpdatedAt = userDB.UpdatedAt

	return nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var userDB User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&userDB).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userDB.ToDomain(), nil
}

func (r *UserRepo) FindAll(ctx context.Context) ([]*domain.User, error) {
	var usersDB []User
	// Gunakan Preload untuk melakukan JOIN ke tabel villages
	if err := r.db.WithContext(ctx).Preload("Village").Find(&usersDB).Error; err != nil {
		return nil, err
	}

	var domainUsers []*domain.User
	for _, u := range usersDB {
		domainUsers = append(domainUsers, u.ToDomain())
	}
	return domainUsers, nil
}

// ----- IMPLEMENTASI BARU ----- //

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var userDB User
	err := r.db.WithContext(ctx).Preload("Village").Where("id = ?", id).First(&userDB).Error
	if err != nil {
		return nil, err
	}
	return userDB.ToDomain(), nil
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	userDB := FromDomain(user)
	// Hanya meng-update field yang tidak kosong
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.ID).Updates(userDB).Error
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	// Fitur Soft Delete bawaan Gorm
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&User{}).Error
}
