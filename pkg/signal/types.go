package signal

import "time"

// TypeStatUpdate is the signal type for generic stat updates (fallback).
const TypeStatUpdate = "stat_update"

// StatUpdateSignal represents a generic statistic update event.
// This is used as a fallback for stat codes that don't have specific signal mappers.
type StatUpdateSignal struct {
	BaseSignal
	StatCode string
	Value    float64
}

// NewStatUpdateSignal creates a new stat update signal.
func NewStatUpdateSignal(userID string, timestamp time.Time, statCode string, value float64, context *PlayerContext) *StatUpdateSignal {
	metadata := map[string]interface{}{
		"stat_code": statCode,
		"value":     value,
	}
	return &StatUpdateSignal{
		BaseSignal: NewBaseSignal(TypeStatUpdate, userID, timestamp, metadata, context),
		StatCode:   statCode,
		Value:      value,
	}
}
