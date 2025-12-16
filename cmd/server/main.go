package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	PasswordHash string `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
}

type UserStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

var store = &UserStore{
	users: make(map[string]*User),
}

func main() {
	http.HandleFunc("/api/auth/register", loggingMiddleware(rateLimitMiddleware(registerHandler)))
	http.HandleFunc("/api/auth/login", loggingMiddleware(rateLimitMiddleware(loginHandler)))
	http.HandleFunc("/api/auth/refresh", loggingMiddleware(refreshTokenHandler))
	http.HandleFunc("/api/auth/forgot-password", loggingMiddleware(rateLimitMiddleware(forgotPasswordHandler)))
	http.HandleFunc("/api/auth/reset-password", loggingMiddleware(rateLimitMiddleware(resetPasswordHandler)))
	http.HandleFunc("/api/auth/verify-email", loggingMiddleware(verifyEmailHandler))
	http.HandleFunc("/api/auth/logout", loggingMiddleware(authMiddleware(logoutHandler)))
	http.HandleFunc("/api/users/me", loggingMiddleware(authMiddleware(meHandler)))
	http.HandleFunc("/api/users/profile", loggingMiddleware(authMiddleware(updateProfileHandler)))
	http.HandleFunc("/health", healthHandler)

	fmt.Println("Go Auth API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "go-auth-api",
	})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !validateEmail(req.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if !validatePassword(req.Password) {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	if _, exists := store.users[req.Email]; exists {
		store.mu.Unlock()
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	
	user := &User{
		ID:          generateID(),
		Email:       req.Email,
		PasswordHash: string(hashedPassword),
		Name:        req.Name,
		CreatedAt:    time.Now(),
	}
	store.users[req.Email] = user
	
	// Generate verification token
	verificationToken := generateVerificationToken(req.Email, user.ID)
	store.mu.Unlock()

	// In production, send verification email
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "User registered successfully. Please verify your email.",
		"user_id":            user.ID,
		"verification_token": verificationToken, // In production, don't return token
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	user, exists := store.users[req.Email]
	store.mu.RUnlock()

	if !exists {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken := generateToken()
	refreshToken := generateToken()
	
	// Store refresh token
	refreshTokens.Store(refreshToken, user.ID, 7*24*time.Hour)
	
	// Create session
	ip := r.RemoteAddr
	userAgent := r.UserAgent()
	sessionStore.Create(user.ID, accessToken, ip, userAgent)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Simplified: In production, validate JWT token
	store.mu.RLock()
	users := make([]*User, 0, len(store.users))
	for _, user := range store.users {
		users = append(users, &User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		})
	}
	store.mu.RUnlock()

	if len(users) > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users[0])
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:16]
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
