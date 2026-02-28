package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type GoldPrices struct {
	Price24K float64 // INR per 10 grams (999 purity)
	Price22K float64 // INR per 10 grams (916 purity)
}

type ibjaData struct {
	Purity999 []float64 `json:"purity999"`
	Purity916 []float64 `json:"purity916"`
}

// FetchGoldPrices fetches live 24K and 22K gold prices in INR per 10 grams
// directly from IBJA (India Bullion and Jewellers Association) — ibjarates.com.
// This is the official Indian bullion source used by jewellers across India.
func FetchGoldPrices() (GoldPrices, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.ibjarates.com/", nil)
	if err != nil {
		return GoldPrices{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GoldPrices{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GoldPrices{}, fmt.Errorf("IBJA returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GoldPrices{}, err
	}

	// Extract the hidden field: <input id="HdnGold" value="{&quot;purity999&quot;:[...]}" />
	re := regexp.MustCompile(`HdnGold[^>]+value="([^"]+)"`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return GoldPrices{}, errors.New("IBJA: could not find gold data in page")
	}

	// Decode HTML entities: &quot; -> "
	jsonStr := strings.ReplaceAll(string(matches[1]), "&quot;", `"`)

	var data ibjaData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return GoldPrices{}, fmt.Errorf("IBJA: failed to parse data: %w", err)
	}

	if len(data.Purity999) == 0 || len(data.Purity916) == 0 {
		return GoldPrices{}, errors.New("IBJA: price data is empty")
	}

	// Last element = today's rate
	return GoldPrices{
		Price24K: data.Purity999[len(data.Purity999)-1],
		Price22K: data.Purity916[len(data.Purity916)-1],
	}, nil
}
