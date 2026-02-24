package main

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"
)

// RateLimiter controls the rate of API calls
type RateLimiter struct {
	minInterval float64
	mu          sync.Mutex
	nextAllowed time.Time
}

// NewRateLimiter creates a new rate limiter with the specified rate per minute
func NewRateLimiter(ratePerMin float64) *RateLimiter {
	if ratePerMin <= 0 {
		ratePerMin = 45.0
	}
	return &RateLimiter{
		minInterval: 60.0 / ratePerMin,
	}
}

// Wait blocks until the next API call is allowed
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	now := time.Now()
	if now.Before(r.nextAllowed) {
		waitTime := r.nextAllowed.Sub(now)
		r.mu.Unlock()

		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	} else {
		r.mu.Unlock()
	}

	r.mu.Lock()
	jitter := rand.Float64() * r.minInterval * 0.1
	r.nextAllowed = time.Now().Add(time.Duration((r.minInterval + jitter) * float64(time.Second)))
	r.mu.Unlock()

	return nil
}

// RetriableError indicates an error that can be retried
type RetriableError struct {
	Message string
}

func (e *RetriableError) Error() string {
	return e.Message
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	MaxAttempts int // 0 means unlimited
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    90 * time.Second,
		MaxAttempts: 0,
	}
}

// Retry executes the given function with exponential backoff retry
func Retry(ctx context.Context, fn func() error, isRetriable func(error) bool, onRetry func(attempt int, err error, delay time.Duration), config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	attempt := 0
	for {
		attempt++
		err := fn()
		if err == nil {
			return nil
		}

		if !isRetriable(err) {
			return err
		}

		if config.MaxAttempts > 0 && attempt >= config.MaxAttempts {
			return err
		}

		delay := time.Duration(math.Min(
			float64(config.MaxDelay),
			float64(config.BaseDelay)*math.Pow(2, float64(attempt-1)),
		))
		delay = time.Duration(float64(delay) * (0.7 + rand.Float64()*0.6))

		if onRetry != nil {
			onRetry(attempt, err, delay)
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// IsRetriableError checks if an error is retriable
func IsRetriableError(err error) bool {
	// Check for explicit retriable error
	var retriable *RetriableError
	if errors.As(err, &retriable) {
		return true
	}

	// Check for context deadline exceeded (timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for temporary errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// Check for specific network errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Check for URL errors (often wraps network errors)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() || urlErr.Temporary() {
			return true
		}
	}

	// Check for syscall errors (connection refused, reset, etc.)
	var syscallErr *syscall.Errno
	if errors.As(err, &syscallErr) {
		switch *syscallErr {
		case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ETIMEDOUT,
			syscall.ENETUNREACH, syscall.EHOSTUNREACH, syscall.EPIPE:
			return true
		}
	}

	// Check error message for common timeout/connection patterns
	errStr := strings.ToLower(err.Error())
	timeoutPatterns := []string{
		"timeout", "deadline exceeded", "connection refused", "connection reset",
		"no such host", "temporary failure", "i/o timeout", "context canceled",
	}
	for _, pattern := range timeoutPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
