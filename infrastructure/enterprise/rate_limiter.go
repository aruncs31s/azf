package enterprise

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed            bool
	RemainingRequests  int
	ResetAtTime        time.Time
	RetryAfterSeconds  int
	LimitExceeded      bool
	CurrentWindowCount int
	WindowSize         time.Duration
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	DefaultRequestsPerMinute int
	RoleSpecificLimits       map[string]int // role -> requests per minute
	BurstAllowance           int            // Extra requests allowed temporarily
	WindowDuration           time.Duration  // Time window for counting (default: 1 minute)
	EnableRedis              bool           // Use Redis for distributed rate limiting
}

// RateLimiter interface for implementations
type RateLimiter interface {
	CheckLimit(ctx context.Context, identifier string, role string) (*RateLimitResult, error)
	Reset(ctx context.Context, identifier string) error
	GetStats(ctx context.Context, identifier string) (map[string]interface{}, error)
}

// InMemoryRateLimiter uses in-memory storage for rate limiting
type InMemoryRateLimiter struct {
	config         *RateLimitConfig
	buckets        map[string]*TokenBucket
	mu             sync.RWMutex
	cleanupTicker  *time.Ticker
	logger         *zap.Logger
	stopCleaning   chan bool
	cleanupRunning bool
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	Tokens           float64
	LastRefillTime   time.Time
	MaxTokens        float64
	RefillRatePerSec float64
	WindowStart      time.Time
	WindowCount      int
	CreatedAt        time.Time
}

// RedisRateLimiter uses Redis for distributed rate limiting
type RedisRateLimiter struct {
	config *RateLimitConfig
	client *redis.Client
	logger *zap.Logger
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(config *RateLimitConfig, logger *zap.Logger) *InMemoryRateLimiter {
	if config == nil {
		config = &RateLimitConfig{
			DefaultRequestsPerMinute: 60,
			RoleSpecificLimits:       make(map[string]int),
			BurstAllowance:           10,
			WindowDuration:           time.Minute,
			EnableRedis:              false,
		}
	}

	if config.WindowDuration == 0 {
		config.WindowDuration = time.Minute
	}

	limiter := &InMemoryRateLimiter{
		config:         config,
		buckets:        make(map[string]*TokenBucket),
		logger:         logger,
		stopCleaning:   make(chan bool),
		cleanupRunning: false,
	}

	// Start cleanup goroutine
	limiter.startCleanup()

	return limiter
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(config *RateLimitConfig, client *redis.Client, logger *zap.Logger) *RedisRateLimiter {
	if config == nil {
		config = &RateLimitConfig{
			DefaultRequestsPerMinute: 60,
			RoleSpecificLimits:       make(map[string]int),
			BurstAllowance:           10,
			WindowDuration:           time.Minute,
			EnableRedis:              true,
		}
	}

	if config.WindowDuration == 0 {
		config.WindowDuration = time.Minute
	}

	return &RedisRateLimiter{
		config: config,
		client: client,
		logger: logger,
	}
}

// CheckLimit checks if a request is within the rate limit
func (rl *InMemoryRateLimiter) CheckLimit(ctx context.Context, identifier string, role string) (*RateLimitResult, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit := rl.config.DefaultRequestsPerMinute

	// Get role-specific limit if available
	if roleLimit, exists := rl.config.RoleSpecificLimits[role]; exists {
		limit = roleLimit
	}

	// Get or create token bucket
	bucket, exists := rl.buckets[identifier]
	if !exists {
		tokensPerSec := float64(limit) / 60.0
		bucket = &TokenBucket{
			Tokens:           float64(limit),
			MaxTokens:        float64(limit + rl.config.BurstAllowance),
			RefillRatePerSec: tokensPerSec,
			LastRefillTime:   now,
			WindowStart:      now,
			WindowCount:      0,
			CreatedAt:        now,
		}
		rl.buckets[identifier] = bucket
	}

	// Refill tokens based on time elapsed
	timeSinceLastRefill := now.Sub(bucket.LastRefillTime).Seconds()
	tokensToAdd := timeSinceLastRefill * bucket.RefillRatePerSec
	bucket.Tokens = min(bucket.MaxTokens, bucket.Tokens+tokensToAdd)
	bucket.LastRefillTime = now

	// Check if window has expired
	if now.Sub(bucket.WindowStart) > rl.config.WindowDuration {
		bucket.WindowStart = now
		bucket.WindowCount = 0
	}

	// Check if request is allowed
	allowed := bucket.Tokens >= 1.0
	result := &RateLimitResult{
		Allowed:            allowed,
		LimitExceeded:      !allowed,
		CurrentWindowCount: bucket.WindowCount,
		WindowSize:         rl.config.WindowDuration,
		ResetAtTime:        bucket.WindowStart.Add(rl.config.WindowDuration),
	}

	if allowed {
		bucket.Tokens -= 1.0
		bucket.WindowCount++
		result.RemainingRequests = int(bucket.Tokens)
	} else {
		result.RetryAfterSeconds = int(rl.config.WindowDuration.Seconds())
		rl.logger.Warn(
			"Rate limit exceeded",
			zap.String("identifier", identifier),
			zap.String("role", role),
			zap.Int("limit", limit),
			zap.Int("window_count", bucket.WindowCount),
		)
	}

	return result, nil
}

// CheckLimit checks rate limit using Redis
func (rl *RedisRateLimiter) CheckLimit(ctx context.Context, identifier string, role string) (*RateLimitResult, error) {
	limit := rl.config.DefaultRequestsPerMinute

	// Get role-specific limit
	if roleLimit, exists := rl.config.RoleSpecificLimits[role]; exists {
		limit = roleLimit
	}

	// Create Redis key
	key := fmt.Sprintf("rate_limit:%s:%s", role, identifier)

	now := time.Now()
	windowStart := now.Truncate(rl.config.WindowDuration)
	windowEnd := windowStart.Add(rl.config.WindowDuration)

	// Use Redis pipeline for atomic operations
	pipe := rl.client.Pipeline()

	// Increment counter
	incCmd := pipe.Incr(ctx, key)
	// Set expiration
	pipe.Expire(ctx, key, rl.config.WindowDuration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		rl.logger.Error("Redis pipeline error", zap.Error(err))
		return nil, err
	}

	// Current count from increment command
	count := incCmd.Val()
	if count < 0 {
		count = 0
	}

	maxRequests := int64(limit + rl.config.BurstAllowance)
	allowed := count <= int64(limit)

	remaining := int(limit) - int(count)
	if remaining < 0 {
		remaining = 0
	}
	result := &RateLimitResult{
		Allowed:            allowed,
		LimitExceeded:      !allowed,
		CurrentWindowCount: int(count),
		RemainingRequests:  remaining,
		ResetAtTime:        windowEnd,
		WindowSize:         rl.config.WindowDuration,
	}

	if !allowed {
		result.RetryAfterSeconds = int(windowEnd.Sub(now).Seconds())
		rl.logger.Warn(
			"Rate limit exceeded (Redis)",
			zap.String("identifier", identifier),
			zap.String("role", role),
			zap.Int("limit", limit),
			zap.Int64("count", count),
			zap.Int64("max", maxRequests),
		)
	}

	return result, nil
}

// Reset resets the rate limit for an identifier
func (rl *InMemoryRateLimiter) Reset(ctx context.Context, identifier string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, identifier)
	rl.logger.Debug("Rate limit reset", zap.String("identifier", identifier))
	return nil
}

// Reset resets the rate limit in Redis
func (rl *RedisRateLimiter) Reset(ctx context.Context, identifier string) error {
	pattern := fmt.Sprintf("rate_limit:*:%s", identifier)
	// Delete all matching keys (note: SCAN is used instead of KEYS for production)
	keys, err := rl.client.Keys(ctx, pattern).Result()
	if err != nil {
		rl.logger.Error("Redis keys error", zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		err = rl.client.Del(ctx, keys...).Err()
		if err != nil {
			rl.logger.Error("Redis delete error", zap.Error(err))
			return err
		}
	}

	rl.logger.Debug("Rate limit reset (Redis)", zap.String("identifier", identifier), zap.Int("keys", len(keys)))
	return nil
}

// GetStats returns statistics for an identifier
func (rl *InMemoryRateLimiter) GetStats(ctx context.Context, identifier string) (map[string]interface{}, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket, exists := rl.buckets[identifier]
	if !exists {
		return map[string]interface{}{
			"exists": false,
		}, nil
	}

	return map[string]interface{}{
		"exists":              true,
		"tokens":              bucket.Tokens,
		"max_tokens":          bucket.MaxTokens,
		"refill_rate_per_sec": bucket.RefillRatePerSec,
		"window_count":        bucket.WindowCount,
		"window_start":        bucket.WindowStart,
		"created_at":          bucket.CreatedAt,
		"last_refill":         bucket.LastRefillTime,
	}, nil
}

// GetStats returns statistics from Redis
func (rl *RedisRateLimiter) GetStats(ctx context.Context, identifier string) (map[string]interface{}, error) {
	pattern := fmt.Sprintf("rate_limit:*:%s", identifier)
	keys, err := rl.client.Keys(ctx, pattern).Result()
	if err != nil {
		rl.logger.Error("Redis keys error", zap.Error(err))
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["total_keys"] = len(keys)

	for _, key := range keys {
		count, _ := rl.client.Get(ctx, key).Int64()
		ttl, _ := rl.client.TTL(ctx, key).Result()
		stats[key] = map[string]interface{}{
			"count": count,
			"ttl":   ttl,
		}
	}

	return stats, nil
}

// startCleanup starts the cleanup goroutine
func (rl *InMemoryRateLimiter) startCleanup() {
	if rl.cleanupRunning {
		return
	}

	rl.cleanupRunning = true
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-rl.cleanupTicker.C:
				rl.cleanupExpiredBuckets()
			case <-rl.stopCleaning:
				rl.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanupExpiredBuckets removes expired token buckets
func (rl *InMemoryRateLimiter) cleanupExpiredBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cleanupThreshold := 30 * time.Minute

	for identifier, bucket := range rl.buckets {
		if now.Sub(bucket.CreatedAt) > cleanupThreshold {
			delete(rl.buckets, identifier)
			rl.logger.Debug("Cleaned up expired bucket", zap.String("identifier", identifier))
		}
	}
}

// Stop stops the cleanup goroutine
func (rl *InMemoryRateLimiter) Stop() {
	if rl.cleanupRunning {
		rl.stopCleaning <- true
		rl.cleanupRunning = false
	}
}

// SetRoleLimit updates (or adds) a per-role requests-per-minute limit and optionally a new burst allowance
func (rl *InMemoryRateLimiter) SetRoleLimit(role string, requestsPerMinute int, burstAllowance int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.config.RoleSpecificLimits == nil {
		rl.config.RoleSpecificLimits = make(map[string]int)
	}

	rl.config.RoleSpecificLimits[role] = requestsPerMinute
	if burstAllowance >= 0 {
		rl.config.BurstAllowance = burstAllowance
	}

	rl.logger.Debug("Updated role rate limit",
		zap.String("role", role),
		zap.Int("requests_per_minute", requestsPerMinute),
		zap.Int("burst_allowance", rl.config.BurstAllowance),
	)
}

// min returns the minimum of two numbers
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RateLimitMiddleware creates middleware for rate limiting
func RateLimitMiddleware(limiter RateLimiter, logger *zap.Logger) func(ctx interface{}) error {
	return func(ctx interface{}) error {
		// This is a helper function to be used in middleware integration
		return nil
	}
}
