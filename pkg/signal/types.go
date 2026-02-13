package signal

import "time"

// TypeStatUpdate is the signal type for generic stat updates (fallback).
const TypeStatUpdate = "stat_update"

// StatUpdateSignal represents a generic statistic update event.
// This is used as a fallback for stat codes that don't have specific signal mappers.
type StatUpdateSignal struct {
	signalType string
	userID     string
	timestamp  time.Time
	metadata   map[string]interface{}
	context    *PlayerContext
	StatCode   string
	Value      float64
}

// NewStatUpdateSignal creates a new stat update signal.
func NewStatUpdateSignal(userID string, timestamp time.Time, statCode string, value float64, context *PlayerContext) *StatUpdateSignal {
	metadata := map[string]interface{}{
		"stat_code": statCode,
		"value":     value,
	}
	return &StatUpdateSignal{
		signalType: TypeStatUpdate,
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
		StatCode:   statCode,
		Value:      value,
	}
}

// Type implements Signal interface.
func (s *StatUpdateSignal) Type() string {
	return s.signalType
}

// UserID implements Signal interface.
func (s *StatUpdateSignal) UserID() string {
	return s.userID
}

// Timestamp implements Signal interface.
func (s *StatUpdateSignal) Timestamp() time.Time {
	return s.timestamp
}

// Metadata implements Signal interface.
func (s *StatUpdateSignal) Metadata() map[string]interface{} {
	return s.metadata
}

// Context implements Signal interface.
func (s *StatUpdateSignal) Context() *PlayerContext {
	return s.context
}
