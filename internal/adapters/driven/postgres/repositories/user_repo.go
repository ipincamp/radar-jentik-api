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

func (r *UserRepo) GetPaginatedUsers(ctx context.Context, page int, limit int) ([]*domain.User, int64, error) {
	var usersDB []User
	var totalData int64

	// 1. Buat Base Query (Bisa ditambahkan .Where("role = ?", "cadre") jika butuh filter khusus kader)
	baseQuery := r.db.WithContext(ctx).Model(&User{})

	// 2. Hitung TOTAL DATA keseluruhan (wajib dipanggil SEBELUM Limit & Offset)
	if err := baseQuery.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	// Jika data kosong, langsung return agar tidak perlu eksekusi database lagi
	if totalData == 0 {
		return []*domain.User{}, 0, nil
	}

	// 3. Hitung OFFSET (Data dilewati)
	// Rumus baku: (Halaman saat ini - 1) * Limit
	// Contoh: Halaman 2, limit 10 -> (2-1) * 10 = 10. Berarti lewati 10 baris pertama.
	offset := (page - 1) * limit

	// 4. Ambil Data (Tarik dari DB dengan Limit, Offset, dan Preload/Join)
	err := baseQuery.
		Preload("Village").       // Tetap lakukan Join ke Master Desa
		Order("created_at DESC"). // Urutkan dari yang terbaru (Best Practice)
		Limit(limit).
		Offset(offset).
		Find(&usersDB).Error

	if err != nil {
		return nil, 0, err
	}

	// 5. Mapping ke Object Domain
	var domainUsers []*domain.User
	for _, u := range usersDB {
		domainUsers = append(domainUsers, u.ToDomain())
	}

	return domainUsers, totalData, nil
}
