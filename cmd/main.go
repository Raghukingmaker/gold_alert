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
	// Change indicator
	changeColor := "#27ae60" // green = price up
	changeSymbol := "&#9650;" // ▲
	if diff < 0 {
		changeColor = "#e74c3c" // red = price down
		changeSymbol = "&#9660;" // ▼
	} else if diff == 0 {
		changeColor = "#888888"
		changeSymbol = "&#9679;" // ●
	}

	prev24K := "—"
	prev22K := "—"
	changeSection := `<tr><td colspan="3" style="padding:12px 20px;font-size:13px;color:#888;text-align:center;">First run — no previous price on record.</td></tr>`

	if state.LastPrice24K > 0 {
		prev24K = "&#8377;" + formatINR(state.LastPrice24K)
		prev22K = "&#8377;" + formatINR(state.LastPrice22K)
		changeSection = fmt.Sprintf(`
		<tr style="background:#1a1a2e;">
			<td style="padding:14px 20px;color:#aaa;font-size:13px;">24K Change</td>
			<td colspan="2" style="padding:14px 20px;font-size:18px;font-weight:bold;color:%s;text-align:right;">
				%s &nbsp;&#8377;%s &nbsp;<span style="font-size:13px;font-weight:normal;">(%.2f%%)</span>
			</td>
		</tr>`,
			changeColor, changeSymbol, formatINR(math.Abs(diff)), math.Abs(pct),
		)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Gold Price Alert</title></head>
<body style="margin:0;padding:0;background-color:#0f0f1a;font-family:Arial,Helvetica,sans-serif;">

<table width="100%%" cellpadding="0" cellspacing="0" style="background:#0f0f1a;padding:30px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;border-radius:12px;overflow:hidden;box-shadow:0 8px 32px rgba(0,0,0,0.5);">

  <!-- Header -->
  <tr>
    <td colspan="3" style="background:linear-gradient(135deg,#b8860b,#ffd700,#b8860b);padding:32px 20px;text-align:center;">
      <div style="font-size:28px;margin-bottom:4px;">&#127947;</div>
      <div style="font-size:22px;font-weight:bold;color:#1a1a00;letter-spacing:2px;">GOLD PRICE ALERT</div>
      <div style="font-size:12px;color:#5a4a00;margin-top:4px;letter-spacing:1px;">INDIA &nbsp;|&nbsp; Rate per 10 Grams (Approx.)</div>
    </td>
  </tr>

  <!-- Column Headers -->
  <tr style="background:#16213e;">
    <td style="padding:12px 20px;font-size:11px;font-weight:bold;color:#ffd700;letter-spacing:1px;text-transform:uppercase;">Purity</td>
    <td style="padding:12px 20px;font-size:11px;font-weight:bold;color:#aaa;letter-spacing:1px;text-transform:uppercase;text-align:right;">Previous</td>
    <td style="padding:12px 20px;font-size:11px;font-weight:bold;color:#ffd700;letter-spacing:1px;text-transform:uppercase;text-align:right;">Current</td>
  </tr>

  <!-- 24K Row -->
  <tr style="background:#0d1b2a;">
    <td style="padding:20px;border-left:4px solid #ffd700;">
      <div style="font-size:20px;font-weight:bold;color:#ffd700;">24K</div>
      <div style="font-size:11px;color:#888;margin-top:2px;">Pure Gold</div>
    </td>
    <td style="padding:20px;text-align:right;color:#666;font-size:15px;">%s</td>
    <td style="padding:20px;text-align:right;">
      <div style="font-size:22px;font-weight:bold;color:#ffd700;">&#8377;%s</div>
    </td>
  </tr>

  <!-- 22K Row -->
  <tr style="background:#111827;">
    <td style="padding:20px;border-left:4px solid #c9a84c;">
      <div style="font-size:20px;font-weight:bold;color:#c9a84c;">22K</div>
      <div style="font-size:11px;color:#888;margin-top:2px;">Jewellery Gold</div>
    </td>
    <td style="padding:20px;text-align:right;color:#666;font-size:15px;">%s</td>
    <td style="padding:20px;text-align:right;">
      <div style="font-size:22px;font-weight:bold;color:#c9a84c;">&#8377;%s</div>
    </td>
  </tr>

  <!-- Change Row -->
  %s

  <!-- Divider -->
  <tr><td colspan="3" style="background:#ffd700;height:2px;padding:0;"></td></tr>

  <!-- Info Section -->
  <tr style="background:#16213e;">
    <td colspan="3" style="padding:20px;">
      <table width="100%%" cellpadding="0" cellspacing="0">
        <tr>
          <td style="padding:6px 0;font-size:12px;color:#aaa;">&#128196; Trigger</td>
          <td style="padding:6px 0;font-size:12px;color:#fff;text-align:right;">%s</td>
        </tr>
        <tr>
          <td style="padding:6px 0;font-size:12px;color:#aaa;">&#128336; Time (IST)</td>
          <td style="padding:6px 0;font-size:12px;color:#fff;text-align:right;">%s</td>
        </tr>
      </table>
    </td>
  </tr>

  <!-- Footer -->
  <tr>
    <td colspan="3" style="background:#0a0a14;padding:16px 20px;text-align:center;font-size:11px;color:#555;line-height:1.6;">
      Prices derived from XAU/USD spot rate converted to INR.<br>
      Actual jewellery prices may vary by city &amp; making charges.
    </td>
  </tr>

</table>
</td></tr>
</table>

</body>
</html>`,
		prev24K, formatINR(prices.Price24K),
		prev22K, formatINR(prices.Price22K),
		changeSection,
		reason,
		now.Format("02 Jan 2006  15:04:05"),
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
