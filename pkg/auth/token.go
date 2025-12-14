package auth

import (
	"errors"
	"time"

	"github.com/ipincamp/radar-jentik-api/pkg/config"
	"github.com/o1egl/paseto"
)

// TokenManager struct menampung konfigurasi agar tidak hardcode
type TokenManager struct {
	paseto      *paseto.V2
	secretKey   []byte
	expDuration time.Duration
	audience    string
	issuer      string
}

// NewTokenManager membuat instance baru dengan config yang disuntikkan
func NewTokenManager(cfg *config.Config) *TokenManager {
	return &TokenManager{
		paseto:      paseto.NewV2(),
		secretKey:   []byte(cfg.PasetoSecret),
		expDuration: cfg.PasetoExp,
		audience:    cfg.PasetoAudience,
		issuer:      cfg.PasetoIssuer,
	}
}

// GenerateToken sekarang menjadi method dari struct TokenManager
func (tm *TokenManager) GenerateToken(userID, role string) (string, error) {
	now := time.Now()
	exp := now.Add(tm.expDuration)

	jsonToken := paseto.JSONToken{
		Audience:   tm.audience,
		Issuer:     tm.issuer,
		Jti:        userID,
		Subject:    userID,
		IssuedAt:   now,
		Expiration: exp,
		NotBefore:  now,
	}

	// Masukkan Custom Claim (Role) ke Footer
	footer := map[string]string{"role": role}

	return tm.paseto.Encrypt(tm.secretKey, jsonToken, footer)
}

// ValidateToken memverifikasi token string dan mengembalikan ID pengguna (Subject) jika valid
func (tm *TokenManager) ValidateToken(tokenString string) (string, error) {
	var jsonToken paseto.JSONToken
	var footer map[string]string

	// 1. Decrypt token menggunakan Secret Key
	if err := tm.paseto.Decrypt(tokenString, tm.secretKey, &jsonToken, &footer); err != nil {
		return "", errors.New("token tidak valid")
	}

	// 2. Validasi Expiration
	if time.Now().After(jsonToken.Expiration) {
		return "", errors.New("token sudah kadaluarsa")
	}

	// 3. Validasi Audience & Issuer (Opsional tapi direkomendasikan)
	if jsonToken.Audience != tm.audience {
		return "", errors.New("token audience tidak valid")
	}
	if jsonToken.Issuer != tm.issuer {
		return "", errors.New("token issuer tidak valid")
	}

	// 4. Ambil User ID (disimpan di field Subject saat Generate)
	userID := jsonToken.Subject
	if userID == "" {
		return "", errors.New("token tidak memiliki subject (user_id)")
	}

	return userID, nil
}
