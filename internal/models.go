package internal

import "time"

type State struct {
	LastPrice24K    float64   `json:"last_price_24k"`
	LastPrice22K    float64   `json:"last_price_22k"`
	LastEmailSentAt time.Time `json:"last_email_sent_at"`
}
