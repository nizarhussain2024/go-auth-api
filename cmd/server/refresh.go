package main

import (
	"encoding/json"
	"net/http"
)

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, valid := refreshTokens.Validate(req.RefreshToken)
	if !valid {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	newAccessToken := generateToken()
	newRefreshToken := generateToken()

	// Store new refresh token
	refreshTokens.Store(newRefreshToken, userID, 7*24*time.Hour)

	// Revoke old refresh token
	refreshTokens.Revoke(req.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Token{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}




