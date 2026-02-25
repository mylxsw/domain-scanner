package model

import (
	"time"
)

// TldMode represents the TLD filtering mode
type TldMode string

const (
	TldModeAll                 TldMode = "all"
	TldModeApiRegisterableOnly TldMode = "api-registerable-only"
	TldModeMainstreamOnly      TldMode = "mainstream-only"
)

// ProbeStatus represents the status of a probe task
type ProbeStatus string

const (
	ProbeStatusPending   ProbeStatus = "pending"
	ProbeStatusRunning   ProbeStatus = "running"
	ProbeStatusCompleted ProbeStatus = "completed"
	ProbeStatusFailed    ProbeStatus = "failed"
)

// ProbeTask represents a domain probe task
type ProbeTask struct {
	ID             string      `json:"id"`
	Word           string      `json:"word"`
	Status         ProbeStatus `json:"status"`
	TldMode        TldMode     `json:"tld_mode"`
	RatePerMin     float64     `json:"rate_per_min"`
	MaxBatch       int         `json:"max_batch"`
	CacheTTLHours  float64     `json:"cache_ttl_hours"`
	Total          int         `json:"total"`
	Completed      int         `json:"completed"`
	Error          string      `json:"error,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	FinishedAt     *time.Time  `json:"finished_at,omitempty"`
	Outdir         string      `json:"outdir,omitempty"`
	ReportMd       string      `json:"report_md,omitempty"`
	ResultsCsv     string      `json:"results_csv,omitempty"`
	ResultsJsonl   string      `json:"results_jsonl,omitempty"`
	ElapsedSeconds float64     `json:"elapsed_seconds,omitempty"`
}

// ProbeItem represents a single domain probe result
type ProbeItem struct {
	Domain               string                 `json:"domain"`
	Tld                  string                 `json:"tld"`
	Available            *bool                  `json:"available"`
	IsPremium            *bool                  `json:"is_premium"`
	Currency             *string                `json:"currency"`
	BaseRegisterPrice    *float64               `json:"base_register_price"`
	PremiumRegisterPrice *float64               `json:"premium_register_price"`
	IcannFee             *float64               `json:"icann_fee"`
	EapFee               *float64               `json:"eap_fee"`
	TotalPrice           *float64               `json:"total_price"`
	Error                *string                `json:"error"`
	Raw                  map[string]interface{} `json:"raw"`
}

// ProgressEvent represents a progress event
type ProgressEvent struct {
	Type  string    `json:"type"`
	Index int       `json:"index"`
	Total int       `json:"total"`
	Data  ProbeItem `json:"data"`
}

// PhaseEvent represents a phase change during probing
type PhaseEvent struct {
	Type    string `json:"type"`
	Phase   string `json:"phase"`
	Message string `json:"message"`
}

// SummaryEvent represents the final summary
type SummaryEvent struct {
	Type           string  `json:"type"`
	Word           string  `json:"word"`
	Total          int     `json:"total"`
	Outdir         string  `json:"outdir"`
	ReportMd       string  `json:"report_md"`
	ResultsCsv     string  `json:"results_csv"`
	ResultsJsonl   string  `json:"results_jsonl"`
	FinishedAt     string  `json:"finished_at"`
	ElapsedSeconds float64 `json:"elapsed_seconds"`
}

// CreateProbeRequest represents the request to create a probe task
type CreateProbeRequest struct {
	Word          string  `json:"word" binding:"required"`
	TldMode       string  `json:"tld_mode,omitempty"`
	RatePerMin    float64 `json:"rate_per_min,omitempty"`
	MaxBatch      int     `json:"max_batch,omitempty"`
	CacheTTLHours float64 `json:"cache_ttl_hours,omitempty"`
}

// TLD represents a top-level domain
type TLD struct {
	Name              string `json:"name"`
	IsApiRegisterable bool   `json:"is_api_registerable"`
	IsActive          bool   `json:"is_active"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	Data  string
}

// String returns the SSE formatted string
func (e *SSEEvent) String() string {
	return "event: " + e.Event + "\ndata: " + e.Data + "\n\n"
}
