package settlement

import (
	"fmt"
	"sync"
	"time"
)

type Event struct {
	Timestamp int64       `json:"timestamp"`
	Status    string      `json:"status"`
	Details   interface{} `json:"details"`
}

type TraceRecord struct {
	TraceID            string      `json:"trace_id"`
	Network            string      `json:"network"`
	Scheme             string      `json:"scheme"`
	Status             string      `json:"status"`
	Timeline           []Event     `json:"timeline"`
	AssignedFacilitator string      `json:"assigned_facilitator,omitempty"`
	SettlementProof    interface{} `json:"settlement_proof,omitempty"`
	CreatedAt          int64       `json:"created_at"`
	UpdatedAt          int64       `json:"updated_at"`
}

type StateMachine struct {
	store map[string]*TraceRecord
	mu    sync.RWMutex
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		store: make(map[string]*TraceRecord),
	}
}

func (sm *StateMachine) CreateTrace(traceID, network, scheme string) *TraceRecord {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().Unix()
	record := &TraceRecord{
		TraceID:   traceID,
		Network:   network,
		Scheme:    scheme,
		Status:    "PENDING",
		CreatedAt: now,
		UpdatedAt: now,
		Timeline: []Event{
			{Timestamp: now, Status: "PENDING", Details: "Trace lifecycle initiated"},
		},
	}
	sm.store[traceID] = record
	return record
}

func (sm *StateMachine) Transition(traceID, newStatus string, details interface{}) (*TraceRecord, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	record, exists := sm.store[traceID]
	if !exists {
		return nil, fmt.Errorf("trace ID %s not found", traceID)
	}

	now := time.Now().Unix()
	record.Status = newStatus
	record.UpdatedAt = now
	record.Timeline = append(record.Timeline, Event{
		Timestamp: now,
		Status:    newStatus,
		Details:   details,
	})

	return record, nil
}

func (sm *StateMachine) GetTrace(traceID string) (*TraceRecord, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	record, exists := sm.store[traceID]
	return record, exists
}

func (sm *StateMachine) ExportAuditLog() []*TraceRecord {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var records []*TraceRecord
	for _, rec := range sm.store {
		records = append(records, rec)
	}
	return records
}

var DefaultStateMachine = NewStateMachine()
