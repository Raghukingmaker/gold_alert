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
	// Change indicator colours
	changeColor  := "#00c853" // green  = price up
	changeBg     := "#003320"
	changeSymbol := "&#9650; UP"
	if diff < 0 {
		changeColor  = "#ff5252" // red = price down
		changeBg     = "#2d0000"
		changeSymbol = "&#9660; DOWN"
	} else if diff == 0 {
		changeColor  = "#90caf9"
		changeBg     = "#0d2a4a"
		changeSymbol = "&#9679; NO CHANGE"
	}

	prev24K := "—"
	prev22K := "—"
	changeSection := `
  <tr>
    <td colspan="2" style="background:#0d1b3e;padding:14px 24px;text-align:center;font-size:13px;color:#546e8a;font-style:italic;">
      First alert &mdash; no previous price on record.
    </td>
  </tr>`

	if state.LastPrice24K > 0 {
		prev24K = "&#8377;" + formatINR(state.LastPrice24K)
		prev22K = "&#8377;" + formatINR(state.LastPrice22K)
		changeSection = fmt.Sprintf(`
  <tr>
    <td colspan="2" style="padding:16px 24px;background:#0a1628;">
      <table width="100%%" cellpadding="0" cellspacing="0">
        <tr>
          <td>
            <span style="display:inline-block;background:%s;color:%s;font-size:11px;font-weight:bold;padding:4px 10px;border-radius:4px;letter-spacing:1px;">%s</span>
            <span style="display:inline-block;margin-left:10px;font-size:22px;font-weight:bold;color:%s;">&#8377;%s</span>
            <span style="display:inline-block;margin-left:6px;font-size:13px;color:%s;">(%.2f%%%%)</span>
          </td>
          <td style="text-align:right;font-size:11px;color:#546e8a;">24K change vs last alert</td>
        </tr>
      </table>
    </td>
  </tr>`,
			changeBg, changeColor, changeSymbol,
			changeColor, formatINR(math.Abs(diff)),
			changeColor, math.Abs(pct),
		)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1.0">
  <title>Gold Price Alert</title>
</head>
<body style="margin:0;padding:0;background-color:#060c21;font-family:Arial,Helvetica,sans-serif;">

<table width="100%%" cellpadding="0" cellspacing="0" style="background:#060c21;padding:32px 16px;">
<tr><td align="center">
<table width="560" cellpadding="0" cellspacing="0" style="max-width:560px;width:100%%;border-radius:16px;overflow:hidden;box-shadow:0 16px 48px rgba(0,0,0,0.6);">

  <!-- ═══ HEADER ═══ -->
  <tr>
    <td colspan="2" style="background:linear-gradient(135deg,#1a237e 0%%,#1565c0 50%%,#0277bd 100%%);padding:36px 24px 28px;text-align:center;">
      <div style="font-size:36px;line-height:1;">&#128176;</div>
      <div style="margin-top:10px;font-size:24px;font-weight:bold;color:#ffffff;letter-spacing:3px;text-transform:uppercase;">Gold Price Alert</div>
      <div style="margin-top:6px;display:inline-block;background:rgba(255,255,255,0.15);border-radius:20px;padding:4px 16px;">
        <span style="font-size:12px;color:#90caf9;letter-spacing:1px;">&#127470;&#127475; INDIA &nbsp;&bull;&nbsp; Per 10 Grams</span>
      </div>
      <div style="margin-top:12px;font-size:12px;color:#64b5f6;">%s</div>
    </td>
  </tr>

  <!-- ═══ SECTION LABEL ═══ -->
  <tr>
    <td colspan="2" style="background:#0d1b3e;padding:10px 24px;">
      <span style="font-size:10px;font-weight:bold;color:#1e88e5;letter-spacing:2px;text-transform:uppercase;">&#9679; Live Rates &nbsp;&mdash;&nbsp; Source: IBJA (ibjarates.com)</span>
    </td>
  </tr>

  <!-- ═══ 24K CARD ═══ -->
  <tr style="background:#0d2246;">
    <td style="padding:22px 24px;border-left:5px solid #1e88e5;width:50%%;">
      <div style="font-size:11px;font-weight:bold;color:#1e88e5;letter-spacing:2px;">24K &nbsp;&#183;&nbsp; 999 PURITY</div>
      <div style="font-size:11px;color:#546e8a;margin-top:2px;">Pure Gold</div>
      <div style="font-size:30px;font-weight:bold;color:#ffffff;margin-top:8px;">&#8377;%s</div>
      <div style="font-size:11px;color:#546e8a;margin-top:4px;">prev: %s</div>
    </td>

  <!-- ═══ 22K CARD ═══ -->
    <td style="padding:22px 24px;border-left:3px solid #0d47a1;background:#0a1e40;">
      <div style="font-size:11px;font-weight:bold;color:#42a5f5;letter-spacing:2px;">22K &nbsp;&#183;&nbsp; 916 PURITY</div>
      <div style="font-size:11px;color:#546e8a;margin-top:2px;">Jewellery Gold</div>
      <div style="font-size:30px;font-weight:bold;color:#e3f2fd;margin-top:8px;">&#8377;%s</div>
      <div style="font-size:11px;color:#546e8a;margin-top:4px;">prev: %s</div>
    </td>
  </tr>

  <!-- ═══ CHANGE ROW ═══ -->
  %s

  <!-- ═══ DIVIDER ═══ -->
  <tr><td colspan="2" style="background:linear-gradient(90deg,#1565c0,#0288d1,#1565c0);height:2px;padding:0;"></td></tr>

  <!-- ═══ META INFO ═══ -->
  <tr style="background:#0d1b3e;">
    <td colspan="2" style="padding:18px 24px;">
      <table width="100%%" cellpadding="0" cellspacing="4">
        <tr>
          <td style="font-size:12px;color:#546e8a;padding:5px 0;">&#128204; Trigger</td>
          <td style="font-size:12px;color:#90caf9;text-align:right;padding:5px 0;">%s</td>
        </tr>
        <tr>
          <td style="font-size:12px;color:#546e8a;padding:5px 0;">&#128336; Time (IST)</td>
          <td style="font-size:12px;color:#90caf9;text-align:right;padding:5px 0;">%s</td>
        </tr>
      </table>
    </td>
  </tr>

  <!-- ═══ FOOTER ═══ -->
  <tr>
    <td colspan="2" style="background:#070e24;padding:16px 24px;text-align:center;border-top:1px solid #0d1b3e;">
      <div style="font-size:11px;color:#37474f;line-height:1.7;">
        Rates sourced from <span style="color:#1e88e5;">IBJA</span> (India Bullion &amp; Jewellers Association).<br>
        Actual prices may vary by city &amp; making charges.
      </div>
    </td>
  </tr>

</table>
</td></tr>
</table>

</body>
</html>`,
		now.Format("02 Jan 2006  |  15:04:05 IST"),
		formatINR(prices.Price24K), prev24K,
		formatINR(prices.Price22K), prev22K,
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
