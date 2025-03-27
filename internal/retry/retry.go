package retry

import (
	"fmt"
	"math/rand"
	"time"
)

// Retry function with exponential backoff
func RetryWithExponentialBackoff(operation func() error, maxRetries int, baseDelay time.Duration) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil // Success
		}

		// Exponential backoff calculation
		sleepTime := baseDelay * (1 << attempt)                    // baseDelay * 2^attempt
		jitter := time.Duration(rand.Int63n(int64(sleepTime / 2))) // Add randomness
		sleepTime = sleepTime + jitter

		fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", attempt+1, err, sleepTime)
		time.Sleep(sleepTime)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, err)
}
