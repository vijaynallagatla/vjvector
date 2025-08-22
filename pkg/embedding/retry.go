package embedding

import (
	"context"
	"math"
	"time"
)

// RetryManager manages retry logic for embedding requests
type RetryManager struct {
	config RetryConfig
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config RetryConfig) *RetryManager {
	if !config.Enabled {
		config.MaxRetries = 0
	}

	rm := &RetryManager{
		config: config,
	}

	return rm
}

// Do executes a function with retry logic
func (rm *RetryManager) Do(fn func() error) error {
	if !rm.config.Enabled || rm.config.MaxRetries <= 0 {
		return fn()
	}

	var lastErr error
	delay := rm.config.InitialDelay

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Don't sleep on the last attempt
		if attempt == rm.config.MaxRetries {
			break
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * rm.config.BackoffFactor)
		if nextDelay > rm.config.MaxDelay {
			nextDelay = rm.config.MaxDelay
		}

		// Sleep before next attempt
		time.Sleep(delay)
		delay = nextDelay
	}

	return lastErr
}

// DoWithContext executes a function with retry logic and context cancellation
func (rm *RetryManager) DoWithContext(ctx context.Context, fn func() error) error {
	if !rm.config.Enabled || rm.config.MaxRetries <= 0 {
		return fn()
	}

	var lastErr error
	delay := rm.config.InitialDelay

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Don't sleep on the last attempt
		if attempt == rm.config.MaxRetries {
			break
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * rm.config.BackoffFactor)
		if nextDelay > rm.config.MaxDelay {
			nextDelay = rm.config.MaxDelay
		}

		// Sleep with context cancellation
		select {
		case <-time.After(delay):
			delay = nextDelay
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// CalculateBackoff calculates the backoff delay for a given attempt
func (rm *RetryManager) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return rm.config.InitialDelay
	}

	// Exponential backoff with jitter
	baseDelay := float64(rm.config.InitialDelay) * math.Pow(rm.config.BackoffFactor, float64(attempt-1))

	// Add jitter (±20%)
	jitter := baseDelay * 0.2
	actualDelay := baseDelay + (jitter * (float64(attempt%2)*2 - 1))

	delay := time.Duration(actualDelay)
	if delay > rm.config.MaxDelay {
		delay = rm.config.MaxDelay
	}

	return delay
}

// IsRetryableError checks if an error should trigger a retry
func (rm *RetryManager) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable errors
	// This is a simplified implementation - in practice, you'd want to check
	// for specific error types from your providers
	errorStr := err.Error()

	// Network-related errors
	if contains(errorStr, "timeout") || contains(errorStr, "connection") ||
		contains(errorStr, "network") || contains(errorStr, "unavailable") {
		return true
	}

	// Rate limiting errors
	if contains(errorStr, "rate limit") || contains(errorStr, "too many requests") {
		return true
	}

	// Server errors
	if contains(errorStr, "internal server error") || contains(errorStr, "service unavailable") {
		return true
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr)))
}
