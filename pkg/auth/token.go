package auth

import (
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
