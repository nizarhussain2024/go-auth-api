package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type PasswordResetToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

var resetTokens = make(map[string]*PasswordResetToken)

func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	user, exists := store.users[req.Email]
	store.mu.RUnlock()

	if !exists {
		// Don't reveal if user exists for security
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Generate reset token
	resetToken := generateToken()
	resetTokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// In production, send email with reset link
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Password reset token generated",
		"token":   resetToken, // In production, don't return token
	})
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resetToken, exists := resetTokens[req.Token]
	if !exists || time.Now().After(resetToken.ExpiresAt) {
		http.Error(w, "Invalid or expired reset token", http.StatusBadRequest)
		return
	}

	if !validatePassword(req.NewPassword) {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Find user and update password
	store.mu.Lock()
	for _, user := range store.users {
		if user.ID == resetToken.UserID {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
			user.PasswordHash = string(hashedPassword)
			break
		}
	}
	store.mu.Unlock()

	// Delete used token
	delete(resetTokens, req.Token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully",
	})
}





