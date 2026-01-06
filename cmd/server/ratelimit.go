package main

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

var rateLimiter = &RateLimiter{
	requests: make(map[string][]time.Time),
	limit:    10,
	window:   time.Minute,
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old requests
	if times, exists := rl.requests[ip]; exists {
		validTimes := []time.Time{}
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[ip] = validTimes
	}

	// Check limit
	if len(rl.requests[ip]) >= rl.limit {
		return false
	}

	// Record request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !rateLimiter.Allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}





