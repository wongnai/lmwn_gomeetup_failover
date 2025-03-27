package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker"
)

type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

func NewCircuitBreaker() *CircuitBreaker {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ServiceCircuitBreaker",
		MaxRequests: 1,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
	})
	return &CircuitBreaker{cb: cb}
}

func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return cb.cb.Execute(req)
}
