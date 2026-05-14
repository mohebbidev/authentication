package domain

import "time"

type TokenHash string

type Token struct {
	Hash      TokenHash
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func NewToken(hash TokenHash, expiresAt time.Time) *Token {
	return &Token{
		Hash:      hash,
		ExpiresAt: expiresAt,
	}
}

func (t *Token) IsExpired(now time.Time) bool {
	return now.After(t.ExpiresAt)
}

func (t *Token) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *Token) Revoke(now time.Time) {
	if t.RevokedAt == nil {
		t.RevokedAt = &now
	}
}

func (t *Token) IsActive(now time.Time) bool {
	return !t.IsExpired(now) && !t.IsRevoked()
}
