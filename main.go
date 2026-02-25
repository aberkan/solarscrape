package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	baseURL   = "https://api.locusenergy.com/v3/sites/3857644/data"
	timezone  = "America/Edmonton"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

type dataPoint struct {
	WhSum float64 `json:"Wh_sum"`
	ID    int     `json:"id"`
	TS    string  `json:"ts"`
}

type response struct {
	StatusCode int         `json:"statusCode"`
	Data       []dataPoint `json:"data"`
	Message    string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func main() {
	token := flag.String("token", "", "Bearer token (required). See README for how to get it from the browser.")
	month := flag.String("month", "", "Month in YYYY-MM format (e.g. 2026-01)")
	year := flag.String("year", "", "Year in YYYY format (e.g. 2026). Use instead of -month for full year.")
	flag.Parse()

	if *token == "" {
		fmt.Fprintln(os.Stderr, "Usage: solarscrape -token <bearer-token> (-month YYYY-MM | -year YYYY)")
		fmt.Fprintln(os.Stderr, "Example: solarscrape -token abc123... -month 2026-01")
		fmt.Fprintln(os.Stderr, "Example: solarscrape -token abc123... -year 2026")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *month != "" && *year != "" {
		fmt.Fprintln(os.Stderr, "Use -month or -year, not both.")
		os.Exit(1)
	}
	if *month == "" && *year == "" {
		fmt.Fprintln(os.Stderr, "Provide -month YYYY-MM or -year YYYY.")
		os.Exit(1)
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid timezone %q: %v\n", timezone, err)
		os.Exit(1)
	}

	now := time.Now().In(loc)
	var start, end time.Time

	if *month != "" {
		start, err = time.ParseInLocation("2006-01", *month, loc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid month %q: use YYYY-MM (e.g. 2026-01)\n", *month)
			os.Exit(1)
		}
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, loc)
		endOfRange := start.AddDate(0, 1, 0).Add(-time.Second)
		if now.Before(endOfRange) {
			end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)
		} else {
			end = endOfRange
		}
	} else {
		start, err = time.ParseInLocation("2006", *year, loc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid year %q: use YYYY (e.g. 2026)\n", *year)
			os.Exit(1)
		}
		start = time.Date(start.Year(), 1, 1, 0, 0, 0, 0, loc)
		endOfRange := time.Date(start.Year(), 12, 31, 23, 59, 59, 0, loc)
		if now.Before(endOfRange) {
			end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)
		} else {
			end = endOfRange
		}
	}

	// API expects naive local time (no offset) when tz is set
	const apiTime = "2006-01-02T15:04:05"
	startStr := start.Format(apiTime)
	endStr := end.Format(apiTime)

	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("start", startStr)
	q.Set("end", endStr)
	q.Set("gran", "daily")
	q.Set("fields", "Wh_sum")
	q.Set("tz", timezone)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Authorization", "Bearer "+*token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read response body: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "API error: HTTP %d %s\n", resp.StatusCode, resp.Status)
		if len(body) > 0 {
			fmt.Fprintf(os.Stderr, "response body: %s\n", body)
		}
		os.Exit(1)
	}

	var apiResp response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Fprintf(os.Stderr, "decode response: %v\n", err)
		fmt.Fprintf(os.Stderr, "response body: %s\n", body)
		os.Exit(1)
	}

	if apiResp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "API error: statusCode %d\n", apiResp.StatusCode)
		if apiResp.Message != "" {
			fmt.Fprintf(os.Stderr, "message: %s\n", apiResp.Message)
		}
		if apiResp.Error != "" {
			fmt.Fprintf(os.Stderr, "error: %s\n", apiResp.Error)
		}
		fmt.Fprintf(os.Stderr, "response body: %s\n", body)
		os.Exit(1)
	}

	// Simple output: Date & Wh_sum (one per line) for piping
	for _, d := range apiResp.Data {
		t, err := time.Parse(time.RFC3339, d.TS)
		if err != nil {
			// fallback: print raw ts
			fmt.Printf("%s\t%.3f\n", d.TS, d.WhSum)
			continue
		}
		fmt.Printf("%s\t%.3f\n", t.Format("2006-01-02"), d.WhSum)
	}
}
