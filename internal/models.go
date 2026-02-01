package internal

import "time"

type State struct {
	LastPrice        float64   `json:"last_price"`
	LastEmailSentAt time.Time `json:"last_email_sent_at"`
}
