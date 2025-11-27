package commands

import (
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// rateLimiter implements a simple token bucket rate limiter per user
type rateLimiter struct {
	mu       sync.RWMutex
	users    map[string]*userLimiter
	capacity int           // Maximum number of tokens
	refill   time.Duration // Time between refills
	window   time.Duration // Time window for rate limiting
}

type userLimiter struct {
	tokens     int
	lastRefill time.Time
	lastReset  time.Time
}

// newRateLimiter creates a new rate limiter with the specified capacity and refill rate
// capacity: maximum number of requests allowed
// window: time window for rate limiting (e.g., 1 minute)
func newRateLimiter(capacity int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		users:    make(map[string]*userLimiter),
		capacity: capacity,
		refill:   window / time.Duration(capacity), // Distribute refills evenly across window
		window:   window,
	}
}

// allow checks if a user is allowed to make a request
// Returns true if allowed, false if rate limited
func (rl *rateLimiter) allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limiter, exists := rl.users[userID]

	if !exists {
		// First request from this user
		rl.users[userID] = &userLimiter{
			tokens:     rl.capacity - 1,
			lastRefill: now,
			lastReset:  now,
		}
		return true
	}

	// Clean up old entries (older than 2x window)
	if now.Sub(limiter.lastReset) > rl.window*2 {
		delete(rl.users, userID)
		return true
	}

	// Refill tokens based on time passed
	timeSinceRefill := now.Sub(limiter.lastRefill)
	if timeSinceRefill >= rl.refill {
		tokensToAdd := int(timeSinceRefill / rl.refill)
		limiter.tokens = min(rl.capacity, limiter.tokens+tokensToAdd)
		limiter.lastRefill = now
	}

	// Check if user has tokens available
	if limiter.tokens > 0 {
		limiter.tokens--
		return true
	}

	// Rate limited
	log.Debug("rate limit exceeded", "user", userID, "window", rl.window)
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Global rate limiter: 10 commands per minute per user
// Discord's rate limits are much higher, but this prevents abuse
var commandRateLimiter = newRateLimiter(10, time.Minute)
