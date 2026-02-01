package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"gold-alert/internal"
)

func main() {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatal("failed to load timezone: ", err)
	}
	now := time.Now().In(loc)

	state, err := internal.LoadState()
	if err != nil {
		log.Fatal(err)
	}

	currentPrice, err := internal.FetchGoldPrice24K()
	if err != nil {
		log.Fatal(err)
	}

	priceDiff := currentPrice - state.LastPrice
	absDiff := math.Abs(priceDiff)

	hoursSinceLastMail := now.Sub(state.LastEmailSentAt).Hours()

	trigger := false
	reason := ""

	if state.LastEmailSentAt.IsZero() {
		trigger = true
		reason = "First run"
	} else if absDiff > 1000 {
		trigger = true
		reason = "Price changed by more than ₹1000"
	} else if hoursSinceLastMail >= 6 {
		trigger = true
		reason = "6-hour interval reached"
	}

	if !trigger {
		log.Println("No alert needed")
		return
	}

	percentage := 0.0
	if state.LastPrice > 0 {
		percentage = (priceDiff / state.LastPrice) * 100
	}

	body := fmt.Sprintf(
		`Gold Price Alert (24K - India)

Previous Price : ₹%.2f
Current Price  : ₹%.2f
Change         : ₹%.2f
Percentage     : %.2f%%

Trigger Reason : %s
Timestamp (IST): %s
`,
		state.LastPrice,
		currentPrice,
		priceDiff,
		percentage,
		reason,
		now.Format("02-Jan-2006 15:04:05"),
	)

	subject := fmt.Sprintf("📢 Gold Price Alert: ₹%.0f (24K)", currentPrice)

	if err := internal.SendEmail(subject, body); err != nil {
		log.Fatal(err)
	}

	state.LastPrice = currentPrice
	state.LastEmailSentAt = now

	if err := internal.SaveState(state); err != nil {
		log.Fatal(err)
	}

	log.Println("Alert sent successfully")
}
