package domain

import "time"

type SessionID string
type UserID string

type Session struct {
	ID        SessionID
	UserID    UserID
	ExpiresAt time.Time
	RevokedAt *time.Time 
}

func NewSession(sid SessionID, uid UserID, expiresAt time.Time) *Session {
	return &Session{
		ID:        sid,
		UserID:    uid,
		ExpiresAt: expiresAt,
	}
}

func (s *Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) Revoke(now time.Time) {
	if s.RevokedAt == nil {
		s.RevokedAt = &now
	}
}

func (s *Session) IsActive(now time.Time) bool {
	return !s.IsExpired(now) && !s.IsRevoked()
}
