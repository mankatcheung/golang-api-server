// Package model defines the data structures and domain types used across the application.
package model

import "time"

// Event represents a generic event published to Kafka.
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// ConversionEvent represents a currency conversion result.
type ConversionEvent struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Rate      float64 `json:"rate"`
	Converted float64 `json:"converted"`
	UserID    int64   `json:"user_id,omitempty"`
}
