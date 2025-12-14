package auth

import (
	"time"

	"github.com/o1egl/paseto"
)

var (
	secretKey = []byte("RAHASIA_DAPUR_JANGAN_DISHARE_YA") // TODO: Pindah ke env variable
	pst       = paseto.NewV2()
)

func GenerateToken(userID, role string) (string, error) {
	now := time.Now()
	exp := now.Add(24 * time.Hour) // TODO: Sesuaikan durasi token

	jsonToken := paseto.JSONToken{
		Audience:   "radar-jentik-app", // TODO: Sesuaikan audience
		Issuer:     "radar-jentik-api", // TODO: Sesuaikan issuer
		Jti:        userID,
		Subject:    userID,
		IssuedAt:   now,
		Expiration: exp,
		NotBefore:  now,
	}

	// Masukkan Custom Claim (Role) ke Footer
	footer := map[string]string{"role": role}

	return pst.Encrypt(secretKey, jsonToken, footer)
}
