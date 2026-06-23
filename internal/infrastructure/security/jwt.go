package security

import (
	"authpractice/internal/domain"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTGenerator struct {
	secret          []byte
	accessTokenTTL  time.Duration
}

func NewJWTGenerator(secret string, accessTokenTTL time.Duration) *JWTGenerator {
	return &JWTGenerator{
		secret:         []byte(secret),
		accessTokenTTL: accessTokenTTL,
	}
}

func (g *JWTGenerator) GenerateAccessToken(userID domain.UserID) (string, error) {
	claims := jwt.MapClaims{
		"sub": string(userID),
		"iat": time.Now().UTC().Unix(),
		"exp": time.Now().UTC().Add(g.accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(g.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// GenerateOpaqueToken creates 32 random bytes, returns the hex raw token
// and its SHA-256 hash for storage. Never store the raw token.
func (g *JWTGenerator) GenerateOpaqueToken() (raw string, hash domain.TokenHash, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate opaque token: %w", err)
	}
	raw = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = domain.TokenHash(hex.EncodeToString(sum[:]))
	return raw, hash, nil
}