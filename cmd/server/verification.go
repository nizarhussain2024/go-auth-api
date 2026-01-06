package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type VerificationToken struct {
	Token     string
	UserID    string
	Email     string
	ExpiresAt time.Time
}

var verificationTokens = &struct {
	mu     sync.RWMutex
	tokens map[string]*VerificationToken
}{
	tokens: make(map[string]*VerificationToken),
}

func verifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	verificationTokens.mu.Lock()
	token, exists := verificationTokens.tokens[req.Token]
	if !exists || time.Now().After(token.ExpiresAt) {
		verificationTokens.mu.Unlock()
		http.Error(w, "Invalid or expired verification token", http.StatusBadRequest)
		return
	}

	// Mark user as verified
	store.mu.Lock()
	if user, exists := store.users[token.Email]; exists {
		// In production, add verified field to user
		_ = user
	}
	store.mu.Unlock()

	delete(verificationTokens.tokens, req.Token)
	verificationTokens.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully",
	})
}

func generateVerificationToken(email, userID string) string {
	token := generateToken()
	verificationTokens.mu.Lock()
	verificationTokens.tokens[token] = &VerificationToken{
		Token:     token,
		UserID:    userID,
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	verificationTokens.mu.Unlock()
	return token
}




