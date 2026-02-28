package main

import (
	"fmt"
	"log"
	"math"
	"strings"
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

	prices, err := internal.FetchGoldPrices()
	if err != nil {
		log.Fatal(err)
	}

	priceDiff := prices.Price24K - state.LastPrice24K
	absDiff := math.Abs(priceDiff)
	hoursSinceLastMail := now.Sub(state.LastEmailSentAt).Hours()

	trigger := false
	reason := ""

	if state.LastEmailSentAt.IsZero() {
		trigger = true
		reason = "First run -- welcome to Gold Alert!"
	} else if absDiff > 1000 {
		trigger = true
		reason = "24K price changed by more than Rs.1,000"
	} else if hoursSinceLastMail >= 6 {
		trigger = true
		reason = "Scheduled 6-hour update"
	}

	if !trigger {
		log.Println("No alert needed")
		return
	}

	percentage := 0.0
	if state.LastPrice24K > 0 {
		percentage = (priceDiff / state.LastPrice24K) * 100
	}

	subject := fmt.Sprintf("Gold Alert: 24K Rs.%s | 22K Rs.%s",
		formatINR(prices.Price24K),
		formatINR(prices.Price22K),
	)

	body := buildEmailBody(state, prices, priceDiff, percentage, reason, now)

	if err := internal.SendEmail(subject, body); err != nil {
		log.Fatal(err)
	}

	state.LastPrice24K = prices.Price24K
	state.LastPrice22K = prices.Price22K
	state.LastEmailSentAt = now

	if err := internal.SaveState(state); err != nil {
		log.Fatal(err)
	}

	log.Println("Alert sent successfully")
}

func buildEmailBody(state *internal.State, prices internal.GoldPrices, diff, pct float64, reason string, now time.Time) string {
	sep := strings.Repeat("=", 50)
	thin := strings.Repeat("-", 50)

	arrow := "(+)"
	if diff < 0 {
		arrow = "(-)"
	} else if diff == 0 {
		arrow = "(=)"
	}

	prev24K := "--"
	prev22K := "--"
	changeRow := "  N/A (first run)"
	if state.LastPrice24K > 0 {
		prev24K = "Rs." + formatINR(state.LastPrice24K)
		prev22K = "Rs." + formatINR(state.LastPrice22K)
		changeRow = fmt.Sprintf("  %s Rs.%s  (%.2f%%)", arrow, formatINR(math.Abs(diff)), math.Abs(pct))
	}

	return fmt.Sprintf(`%s
        GOLD PRICE ALERT -- INDIA
       Rate per 10 Grams (Approx.)
%s

  Purity    Previous Price    Current Price
  ------    --------------    -------------
  24 K      %-16s  Rs.%s
  22 K      %-16s  Rs.%s

%s
  Change    :%s
  Trigger   : %s
  Time (IST): %s
%s

  * Prices sourced from international spot
    rates (XAU/USD) converted to INR.
    Actual jewellery prices may vary by
    city and making charges.
%s
`,
		sep,
		sep,
		prev24K, formatINR(prices.Price24K),
		prev22K, formatINR(prices.Price22K),
		thin,
		changeRow,
		reason,
		now.Format("02-Jan-2006  15:04:05"),
		thin,
		sep,
	)
}

// formatINR formats a number in Indian style: 1,54,619
func formatINR(amount float64) string {
	n := int(math.Round(amount))
	if n < 0 {
		return "-" + formatINR(-amount)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	result := s[len(s)-3:]
	s = s[:len(s)-3]
	for len(s) > 2 {
		result = s[len(s)-2:] + "," + result
		s = s[:len(s)-2]
	}
	if len(s) > 0 {
		result = s + "," + result
	}
	return result
}
