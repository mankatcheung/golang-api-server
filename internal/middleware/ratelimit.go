package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter holds per-IP token-bucket limiters and periodically evicts stale entries.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*ipLimiter
	r        rate.Limit
	burst    int
}

// NewRateLimiter creates a RateLimiter with the given steady-state rate (req/s) and burst size.
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*ipLimiter),
		r:        r,
		burst:    burst,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) get(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[ip]
	if !ok {
		l := rate.NewLimiter(rl.r, rl.burst)
		rl.visitors[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
		return l
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupLoop evicts IPs not seen for 10 minutes, running every 5 minutes.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit returns a Gin middleware that enforces the per-IP rate limit.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.get(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
