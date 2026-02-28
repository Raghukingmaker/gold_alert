package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type goldAPIResponse struct {
	Price float64 `json:"price"`
}

type currencyAPIResponse struct {
	USD map[string]float64 `json:"usd"`
}

// FetchGoldPrice24K returns the current 24K gold price in INR per 10 grams.
// It fetches the USD spot price from gold-api.com and converts using live USD/INR rate.
func FetchGoldPrice24K() (float64, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	goldPriceUSD, err := fetchGoldPriceUSD(client)
	if err != nil {
		return 0, fmt.Errorf("fetching gold price: %w", err)
	}

	usdToINR, err := fetchUSDToINR(client)
	if err != nil {
		return 0, fmt.Errorf("fetching exchange rate: %w", err)
	}

	// gold-api.com returns price per troy ounce in USD
	// 1 troy ounce = 31.1035 grams
	// India quotes gold price per 10 grams
	const gramsPerTroyOunce = 31.1035
	pricePerGramUSD := goldPriceUSD / gramsPerTroyOunce
	pricePerTenGramINR := pricePerGramUSD * 10 * usdToINR

	return pricePerTenGramINR, nil
}

func fetchGoldPriceUSD(client *http.Client) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.gold-api.com/price/XAU", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("gold API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result goldAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if result.Price <= 0 {
		return 0, fmt.Errorf("invalid gold price received: %f", result.Price)
	}

	return result.Price, nil
}

func fetchUSDToINR(client *http.Client) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://latest.currency-api.pages.dev/v1/currencies/usd.json", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("currency API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result currencyAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	rate, ok := result.USD["inr"]
	if !ok || rate <= 0 {
		return 0, fmt.Errorf("invalid INR rate received: %f", rate)
	}

	return rate, nil
}
