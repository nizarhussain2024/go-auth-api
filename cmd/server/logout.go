package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Delete session
	sessionStore.Delete(token)

	// Revoke refresh token if provided
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
		if req.RefreshToken != "" {
			refreshTokens.Revoke(req.RefreshToken)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}




