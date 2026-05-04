package api

import "time"

type RegisterDeviceRequest struct {
	APNsToken string `json:"apns_token"`
}

type RegisterDeviceResponse struct {
	DeviceID string `json:"device_id"`
}

type CreateSubscriptionRequest struct {
	Query          string `json:"query"`
	Type           string `json:"type"`
	CadenceSeconds int    `json:"cadence_seconds"`
}

type SubscriptionDTO struct {
	ID             string     `json:"id"`
	Query          string     `json:"query"`
	Type           string     `json:"type"`
	CadenceSeconds int        `json:"cadence_seconds"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	NextRunAt      time.Time  `json:"next_run_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type SignalDTO struct {
	ID             string     `json:"id"`
	SubscriptionID string     `json:"subscription_id"`
	Title          string     `json:"title"`
	Body           *string    `json:"body,omitempty"`
	URL            *string    `json:"url,omitempty"`
	OccursAt       *time.Time `json:"occurs_at,omitempty"`
	SourceDomains  []string   `json:"source_domains"`
	Confidence     float32    `json:"confidence"`
	FirstSeenAt    time.Time  `json:"first_seen_at"`
}

type RunResponse struct {
	NewSignals int `json:"new_signals"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
