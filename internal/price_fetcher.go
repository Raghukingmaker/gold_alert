package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func FetchGoldPrice24K() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.goodreturns.in/gold-rates/", nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	html := string(bodyBytes)

	// Example match: "24K Gold ₹74,230"
	re := regexp.MustCompile(`24K Gold.*?₹\s?([\d,]+)`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return 0, errors.New("unable to parse gold price")
	}

	priceStr := strings.ReplaceAll(matches[1], ",", "")
	return strconv.ParseFloat(priceStr, 64)
}
