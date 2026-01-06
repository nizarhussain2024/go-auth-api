package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type UpdateProfileRequest struct {
	Email    string `json:"email,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

func updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	session, exists := sessionStore.Get(token)
	if !exists {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	users.mu.Lock()
	user, exists := users.users[session.UserID]
	if !exists {
		users.mu.Unlock()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}

	users.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}



