package registry

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
)

type Capabilities struct {
	Networks            []string `json:"networks"`
	Schemes             []string `json:"schemes"`
	MaxAmount           *float64 `json:"max_amount"`
	SupportsIdempotency bool     `json:"supports_idempotency"`
	SupportsReceipt     bool     `json:"supports_receipt"`
	IsCustodial         bool     `json:"is_custodial"`
}

type RollingMetrics struct {
	latencies   []float64
	errors      []bool
	windowSize  int
	TotalReqs   int
	TotalErrors int
	LastHealth  bool
	LastLatency float64
	mu          sync.RWMutex
}

func NewRollingMetrics(window int) *RollingMetrics {
	return &RollingMetrics{
		latencies:  make([]float64, 0, window),
		errors:     make([]bool, 0, window),
		windowSize: window,
	}
}

func (rm *RollingMetrics) Record(latencyMs float64, err bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(rm.latencies) >= rm.windowSize {
		rm.latencies = rm.latencies[1:]
		rm.errors = rm.errors[1:]
	}
	rm.latencies = append(rm.latencies, latencyMs)
	rm.errors = append(rm.errors, err)

	rm.TotalReqs++
	if err {
		rm.TotalErrors++
	}
}

func (rm *RollingMetrics) RecordHealth(ok bool, latencyMs float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.LastHealth = ok
	rm.LastLatency = latencyMs
}

func (rm *RollingMetrics) P50() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if len(rm.latencies) == 0 {
		return 0
	}
	sorted := make([]float64, len(rm.latencies))
	copy(sorted, rm.latencies)
	sort.Float64s(sorted)
	return sorted[len(sorted)/2]
}

func (rm *RollingMetrics) ErrorRate() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if len(rm.errors) == 0 {
		return 0
	}
	errCount := 0
	for _, err := range rm.errors {
		if err {
			errCount++
		}
	}
	return float64(errCount) / float64(len(rm.errors))
}

type RoutingRule struct {
	Network   string   `json:"network"`
	Prefer    string   `json:"prefer"`
	Scheme    string   `json:"scheme"`
	MaxAmount *float64 `json:"max_amount,omitempty"`
}

func (rr *RoutingRule) Matches(network, scheme string, amount float64) bool {
	if rr.Network != "" && rr.Network != network {
		return false
	}
	if rr.Scheme != "" && rr.Scheme != scheme {
		return false
	}
	if rr.MaxAmount != nil && amount > *rr.MaxAmount {
		return false
	}
	return true
}

type RegistryEntry struct {
	Facilitator  BaseFacilitator
	Priority     int
	Capabilities Capabilities
	Breaker      *CircuitBreaker
	Metrics      *RollingMetrics
}

type Engine struct {
	entries []RegistryEntry
	rules   []RoutingRule
	mu      sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		entries: make([]RegistryEntry, 0),
		rules:   make([]RoutingRule, 0),
	}
}

func (e *Engine) Register(f BaseFacilitator, priority int, caps Capabilities) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry := RegistryEntry{
		Facilitator:  f,
		Priority:     priority,
		Capabilities: caps,
		Breaker:      NewCircuitBreaker(3, 20.0),
		Metrics:      NewRollingMetrics(50),
	}
	e.entries = append(e.entries, entry)
	fmt.Printf("[Registry] Registered: %s (network=%s, priority=%d, sim=%v)\n",
		f.Name(), f.Network(), priority, f.IsSimulator())
}

func (e *Engine) AddRoutingRule(rule RoutingRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
	fmt.Printf("[Registry] Routing rule: %s -> prefer '%s'\n", rule.Network, rule.Prefer)
}

func (e *Engine) Entries() []RegistryEntry {
	e.mu.RLock()
	defer e.mu.RUnlock()
	cp := make([]RegistryEntry, len(e.entries))
	copy(cp, e.entries)
	return cp
}

func (e *Engine) Select(ctx context.Context, network string, scores map[string]float64) (BaseFacilitator, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var candidates []RegistryEntry
	for _, entry := range e.entries {
		if entry.Facilitator.Network() == network {
			candidates = append(candidates, entry)
		}
	}

	if scores != nil {
		sort.Slice(candidates, func(i, j int) bool {
			scoreI := scores[candidates[i].Facilitator.Name()]
			scoreJ := scores[candidates[j].Facilitator.Name()]
			if scoreI != scoreJ {
				return scoreI > scoreJ
			}
			return candidates[i].Priority < candidates[j].Priority
		})
	} else {
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Priority < candidates[j].Priority
		})
	}

	// Try real facilitators first
	for _, entry := range candidates {
		if !entry.Facilitator.IsSimulator() && entry.Breaker.CanExecute() {
			return entry.Facilitator, nil
		}
	}

	// Fallback to simulators
	for _, entry := range candidates {
		if entry.Facilitator.IsSimulator() {
			return entry.Facilitator, nil
		}
	}

	return nil, fmt.Errorf("no available facilitator for network=%s", network)
}

func (e *Engine) GetStatus() []map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var status []map[string]interface{}
	for _, entry := range e.entries {
		status = append(status, map[string]interface{}{
			"name":         entry.Facilitator.Name(),
			"network":      entry.Facilitator.Network(),
			"is_simulator": entry.Facilitator.IsSimulator(),
			"priority":     entry.Priority,
			"capabilities": entry.Capabilities,
			"circuit":      string(entry.Breaker.GetState()),
			"metrics": map[string]interface{}{
				"p50_ms":        entry.Metrics.P50(),
				"error_rate":    math.Round(entry.Metrics.ErrorRate()*10000) / 100,
				"total_reqs":    entry.Metrics.TotalReqs,
				"last_health":   entry.Metrics.LastHealth,
				"last_latency":  entry.Metrics.LastLatency,
			},
		})
	}
	return status
}
