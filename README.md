# Gold Price Alert

Automated gold price monitoring service that sends email alerts when 24K gold prices in India change significantly.

## Overview

This service:
- Fetches current 24K gold prices from [goodreturns.in](https://www.goodreturns.in/gold-rates/)
- Sends email alerts via Gmail SMTP
- Runs automatically every hour via GitHub Actions
- Persists state between runs to track price changes

## Alert Triggers

An email alert is sent when any of these conditions are met:

| Condition | Description |
|-----------|-------------|
| First Run | Always sends alert on initial execution |
| Price Change | Gold price changed by more than ₹1000 |
| Time Interval | 6+ hours since last email |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Actions                           │
│                  (Hourly Cron Job)                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    main.go                                  │
│              (Orchestration Logic)                          │
└─────┬─────────────────┬─────────────────┬───────────────────┘
      │                 │                 │
      ▼                 ▼                 ▼
┌───────────┐   ┌──────────────┐   ┌─────────────┐
│ state.json│   │ price_fetcher│   │  notifier   │
│  (State)  │   │ (Web Scrape) │   │   (Email)   │
└───────────┘   └──────┬───────┘   └──────┬──────┘
                       │                  │
                       ▼                  ▼
              ┌────────────────┐   ┌─────────────┐
              │ goodreturns.in │   │ Gmail SMTP  │
              └────────────────┘   └─────────────┘
```

## Project Structure

```
gold_alert/
├── .github/
│   └── workflows/
│       └── gold_alert.yml    # GitHub Actions workflow
├── cmd/
│   └── main.go               # Application entry point
├── internal/
│   ├── models.go             # Data structures (State)
│   ├── price_fetcher.go      # Gold price web scraping
│   ├── notifier.go           # Email notification logic
│   └── state_store.go        # State persistence (JSON)
├── .env.sample               # Environment variables template
├── state.json                # Persistent state file
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## Prerequisites

- Go 1.22 or later
- Gmail account with App Password enabled
- GitHub repository (for automated runs)

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `EMAIL_FROM` | Yes | Gmail sender address |
| `EMAIL_TO` | Yes | Recipient email address |
| `EMAIL_APP_PASSWORD` | Yes | Gmail App Password (not regular password) |

### Setting up Gmail App Password

1. Go to [Google Account Security](https://myaccount.google.com/security)
2. Enable 2-Step Verification (required for App Passwords)
3. Go to [App Passwords](https://myaccount.google.com/apppasswords)
4. Select "Mail" and your device
5. Click "Generate"
6. Copy the 16-character password (use this for `EMAIL_APP_PASSWORD`)

> **Note**: Regular Gmail passwords will not work. You must use an App Password.

## Local Development

### 1. Clone the repository

```bash
git clone https://github.com/your-username/gold_alert.git
cd gold_alert
```

### 2. Set up environment variables

```bash
cp .env.sample .env
# Edit .env with your actual values
```

### 3. Export environment variables

```bash
export EMAIL_FROM="your-email@gmail.com"
export EMAIL_TO="recipient@example.com"
export EMAIL_APP_PASSWORD="your-app-password"
```

### 4. Run the application

```bash
go run ./cmd/main.go
```

## GitHub Actions Deployment

### 1. Add Repository Secrets

Go to your repository → Settings → Secrets and variables → Actions → New repository secret

Add these secrets:

| Secret Name | Value |
|-------------|-------|
| `EMAIL_FROM` | Your Gmail address |
| `EMAIL_TO` | Recipient email address |
| `EMAIL_APP_PASSWORD` | Gmail App Password |

### 2. Enable GitHub Actions

The workflow runs automatically:
- **Schedule**: Every hour at minute 0 (`0 * * * *`)
- **Manual**: Click "Run workflow" in Actions tab

### 3. Verify Workflow

1. Go to Actions tab in your repository
2. Click on "Gold Price Alert" workflow
3. Check run history and logs

## State Management

The service maintains state in `state.json`:

```json
{
  "last_price": 74230.00,
  "last_email_sent_at": "2024-01-15T10:30:00+05:30"
}
```

- **last_price**: Previous gold price (in INR)
- **last_email_sent_at**: Timestamp of last email sent (IST)

GitHub Actions automatically commits state changes after each run.

## Email Alert Format

```
Subject: 📢 Gold Price Alert: ₹74230 (24K)

Gold Price Alert (24K - India)

Previous Price : ₹73000.00
Current Price  : ₹74230.00
Change         : ₹1230.00
Percentage     : 1.68%

Trigger Reason : Price changed by more than ₹1000
Timestamp (IST): 15-Jan-2024 10:30:00
```

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| `EMAIL_FROM environment variable is not set` | Missing env var | Set `EMAIL_FROM` in secrets/environment |
| `EMAIL_APP_PASSWORD environment variable is not set` | Missing env var | Set `EMAIL_APP_PASSWORD` in secrets/environment |
| `unexpected status code: 403` | Website blocking requests | Check if goodreturns.in is accessible |
| `unable to parse gold price` | Website HTML structure changed | Update regex in `price_fetcher.go` |
| `failed to load timezone` | Invalid timezone | Ensure `Asia/Kolkata` is valid on system |
| Authentication failed | Wrong Gmail credentials | Verify App Password is correct |

### Checking GitHub Actions Logs

1. Go to repository → Actions
2. Click on failed workflow run
3. Expand "Run gold alert" step
4. Check error messages

### Testing Locally

```bash
# Test with verbose output
go run ./cmd/main.go 2>&1

# Check if price fetching works
curl -s https://www.goodreturns.in/gold-rates/ | grep -o "24K Gold.*₹[0-9,]*"
```

## Customization

### Change Alert Thresholds

Edit `cmd/main.go`:

```go
// Price threshold (default: ₹1000)
} else if absDiff > 1000 {

// Time threshold (default: 6 hours)
} else if hoursSinceLastMail >= 6 {
```

### Change Schedule

Edit `.github/workflows/gold_alert.yml`:

```yaml
schedule:
  - cron: "0 * * * *"  # Every hour
  # - cron: "0 */2 * * *"  # Every 2 hours
  # - cron: "0 9,18 * * *"  # At 9 AM and 6 PM
```

### Add Multiple Recipients

Edit `internal/notifier.go` to support comma-separated emails:

```go
recipients := strings.Split(to, ",")
return smtp.SendMail(..., recipients, ...)
```

## Dependencies

This project uses only Go standard library - no external dependencies.

## License

MIT License
