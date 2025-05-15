package circuitbreaker

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

func NewCircuitBreaker() *CircuitBreaker {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ServiceCircuitBreaker",
		MaxRequests: 3,                // Allowed in half-open state
		Interval:    10 * time.Second, // Rolling window to reset counts
		Timeout:     5 * time.Second,  // Duration to wait before switching from open to half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip the breaker if the failure ratio is > 60% over 5+ requests
			failRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failRatio >= 0.6
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			fmt.Printf("Circuit %s changed state from %s to %s\n", name, from.String(), to.String())
		},
	})
	return &CircuitBreaker{cb: cb}
}

func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return cb.cb.Execute(req)
}
