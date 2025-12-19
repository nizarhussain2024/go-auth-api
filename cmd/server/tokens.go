package main

import (
	"sync"
	"time"
)

type RefreshTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*TokenInfo
}

type TokenInfo struct {
	UserID    string
	ExpiresAt time.Time
}

var refreshTokens = &RefreshTokenStore{
	tokens: make(map[string]*TokenInfo),
}

func (rts *RefreshTokenStore) Store(token, userID string, expiresIn time.Duration) {
	rts.mu.Lock()
	defer rts.mu.Unlock()
	rts.tokens[token] = &TokenInfo{
		UserID:    userID,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func (rts *RefreshTokenStore) Validate(token string) (string, bool) {
	rts.mu.RLock()
	defer rts.mu.RUnlock()
	info, exists := rts.tokens[token]
	if !exists {
		return "", false
	}
	if time.Now().After(info.ExpiresAt) {
		delete(rts.tokens, token)
		return "", false
	}
	return info.UserID, true
}

func (rts *RefreshTokenStore) Revoke(token string) {
	rts.mu.Lock()
	defer rts.mu.Unlock()
	delete(rts.tokens, token)
}




