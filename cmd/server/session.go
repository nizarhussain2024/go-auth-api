package main

import (
	"sync"
	"time"
)

type Session struct {
	UserID    string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

var sessionStore = &SessionStore{
	sessions: make(map[string]*Session),
}

func (ss *SessionStore) Create(userID, token, ip, userAgent string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.sessions[token] = &Session{
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: ip,
		UserAgent: userAgent,
	}
}

func (ss *SessionStore) Get(token string) (*Session, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	session, exists := ss.sessions[token]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

func (ss *SessionStore) Delete(token string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, token)
}

func (ss *SessionStore) GetUserSessions(userID string) []*Session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	sessions := []*Session{}
	for _, session := range ss.sessions {
		if session.UserID == userID && time.Now().Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}
	return sessions
}




