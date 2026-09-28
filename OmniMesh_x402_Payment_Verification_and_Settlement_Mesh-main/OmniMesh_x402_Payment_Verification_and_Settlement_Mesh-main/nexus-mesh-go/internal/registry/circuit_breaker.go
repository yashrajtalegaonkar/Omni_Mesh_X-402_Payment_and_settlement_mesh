package registry

import (
	"sync"
	"time"
)

type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"
	StateOpen     CircuitState = "OPEN"
	StateHalfOpen CircuitState = "HALF_OPEN"
)

type CircuitBreaker struct {
	failureThreshold int
	recoveryTimeout  time.Duration
	failureCount     int
	state            CircuitState
	lastStateChange  time.Time
	mu               sync.Mutex
}

func NewCircuitBreaker(threshold int, timeoutSeconds float64) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: threshold,
		recoveryTimeout:  time.Duration(timeoutSeconds * float64(time.Second)),
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == StateOpen {
		if now.Sub(cb.lastStateChange) > cb.recoveryTimeout {
			cb.state = StateHalfOpen
			cb.lastStateChange = now
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.state = StateClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	if cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
