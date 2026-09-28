package orchestrator

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"nexus-mesh-go/internal/registry"
)

const (
	ReliabilityWeight = 0.45
	LatencyWeight     = 0.35
	HealthWeight      = 0.20
	MaxLatencyMs      = 2000.0
	SimulatorCap      = 0.30
)

type ScoreSnapshot struct {
	Name           string                 `json:"name"`
	Network        string                 `json:"network"`
	IsSimulator    bool                   `json:"is_simulator"`
	CompositeScore float64                `json:"composite_score"`
	Breakdown      map[string]float64     `json:"breakdown"`
	Metrics        map[string]float64     `json:"metrics"`
	CircuitState   string                 `json:"circuit_state"`
	SelectionCount int                    `json:"selection_count"`
	LastSelected   *time.Time             `json:"last_selected"`
}

type SelectionDecision struct {
	Timestamp      time.Time `json:"ts"`
	Network        string    `json:"network"`
	Selected       string    `json:"selected"`
	Score          float64   `json:"score"`
	RunnerUp       string    `json:"runner_up,omitempty"`
	RunnerUpScore  float64   `json:"runner_up_score,omitempty"`
}

type FacilitatorOrchestrator struct {
	engine     *registry.Engine
	interval   time.Duration
	scores     map[string]*ScoreSnapshot
	decisions  []SelectionDecision
	cancelFunc context.CancelFunc
	mu         sync.RWMutex
}

func NewOrchestrator(eng *registry.Engine, intervalSeconds float64) *FacilitatorOrchestrator {
	return &FacilitatorOrchestrator{
		engine:    eng,
		interval:  time.Duration(intervalSeconds * float64(time.Second)),
		scores:    make(map[string]*ScoreSnapshot),
		decisions: make([]SelectionDecision, 0),
	}
}

func (o *FacilitatorOrchestrator) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	o.cancelFunc = cancel
	go o.runLoop(ctx)
	fmt.Printf("[Orchestrator] Started background scoring (interval=%v)\n", o.interval)
}

func (o *FacilitatorOrchestrator) Stop() {
	if o.cancelFunc != nil {
		o.cancelFunc()
	}
}

func (o *FacilitatorOrchestrator) runLoop(ctx context.Context) {
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	// Initial run immediately
	o.runScoring(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.runScoring(ctx)
		}
	}
}

func (o *FacilitatorOrchestrator) runScoring(ctx context.Context) {
	entries := o.engine.Entries()
	var wg sync.WaitGroup

	type pingResult struct {
		entry     registry.RegistryEntry
		ok        bool
		latencyMs float64
	}

	resChan := make(chan pingResult, len(entries))

	for _, entry := range entries {
		wg.Add(1)
		go func(e registry.RegistryEntry) {
			defer wg.Done()
			start := time.Now()
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			ok, _ := e.Facilitator.CheckHealth(pingCtx)
			latency := float64(time.Since(start).Microseconds()) / 1000.0
			e.Metrics.RecordHealth(ok, latency)

			if ok {
				e.Breaker.RecordSuccess()
			} else {
				e.Breaker.RecordFailure()
			}

			resChan <- pingResult{entry: e, ok: ok, latencyMs: latency}
		}(entry)
	}

	wg.Wait()
	close(resChan)

	o.mu.Lock()
	defer o.mu.Unlock()

	for res := range resChan {
		e := res.entry
		m := e.Metrics
		cb := e.Breaker

		var score, reliability, latencySc, healthBonus float64

		if cb.GetState() == registry.StateOpen {
			score = 0
		} else {
			reliability = 1.0 - m.ErrorRate()
			p50 := m.P50()
			if p50 > 0 {
				latencySc = 1.0 - math.Min(p50/MaxLatencyMs, 1.0)
			} else {
				latencySc = 0.5
			}
			if res.ok {
				healthBonus = 1.0
			}

			score = ReliabilityWeight*reliability + LatencyWeight*latencySc + HealthWeight*healthBonus
			if e.Facilitator.IsSimulator() {
				score = math.Min(score, SimulatorCap)
			}
		}

		existing, found := o.scores[e.Facilitator.Name()]
		var selCount int
		var lastSel *time.Time
		if found {
			selCount = existing.SelectionCount
			lastSel = existing.LastSelected
		}

		o.scores[e.Facilitator.Name()] = &ScoreSnapshot{
			Name:           e.Facilitator.Name(),
			Network:        e.Facilitator.Network(),
			IsSimulator:    e.Facilitator.IsSimulator(),
			CompositeScore: math.Round(score*10000) / 10000,
			Breakdown: map[string]float64{
				"reliability":   math.Round(reliability*10000) / 10000,
				"latency_score": math.Round(latencySc*10000) / 10000,
				"health_bonus":  healthBonus,
			},
			Metrics: map[string]float64{
				"p50_ms":     m.P50(),
				"error_rate": m.ErrorRate(),
			},
			CircuitState:   string(cb.GetState()),
			SelectionCount: selCount,
			LastSelected:   lastSel,
		}
	}
}

func (o *FacilitatorOrchestrator) GetScoresMap(network string) map[string]float64 {
	o.mu.RLock()
	defer o.mu.RUnlock()

	result := make(map[string]float64)
	for _, snap := range o.scores {
		if network == "" || snap.Network == network {
			result[snap.Name] = snap.CompositeScore
		}
	}
	return result
}

func (o *FacilitatorOrchestrator) GetScores() []*ScoreSnapshot {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var list []*ScoreSnapshot
	for _, snap := range o.scores {
		list = append(list, snap)
	}
	return list
}

func (o *FacilitatorOrchestrator) GetDecisions() []SelectionDecision {
	o.mu.RLock()
	defer o.mu.RUnlock()
	cp := make([]SelectionDecision, len(o.decisions))
	copy(cp, o.decisions)
	return cp
}
