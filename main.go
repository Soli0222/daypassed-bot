package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	dateFormat = "2006-01-02" // YYYY-MM-DD
)

// MisskeyNotePayload defines the structure for the Misskey API request.
type MisskeyNotePayload struct {
	Token string `json:"i"`
	Text  string `json:"text"`
}

type config struct {
	specificDate  time.Time
	mkToken       string
	misskeyAPIURL string
	customText    string
	location      *time.Location
}

func buildAPIURL(host string) string {
	scheme := "https"
	if strings.Contains(host, "localhost") {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/api/notes/create", scheme, host)
}

func calculateDaysPassed(specificDate, now time.Time) float64 {
	timeDifference := now.Sub(specificDate)
	daysPassed := math.Floor(timeDifference.Hours() / 24)
	if daysPassed < 0 {
		return 0
	}
	return daysPassed
}

func formatNoteText(customText string, daysPassed float64) string {
	return fmt.Sprintf("<center>%s\n\n$[jelly $[sparkle %.0f日]]</center>", customText, daysPassed)
}

func loadConfig() (*config, error) {
	specificDateStr := os.Getenv("SPECIFIC_DATE")
	if specificDateStr == "" {
		return nil, fmt.Errorf("SPECIFIC_DATE environment variable not set. Expected format: YYYY-MM-DD")
	}

	mkToken := os.Getenv("MK_TOKEN")
	if mkToken == "" {
		return nil, fmt.Errorf("MK_TOKEN environment variable not set")
	}

	misskeyHost := os.Getenv("MISSKEY_HOST")
	if misskeyHost == "" {
		return nil, fmt.Errorf("MISSKEY_HOST environment variable not set")
	}

	customText := os.Getenv("CUSTOM_TEXT")
	if customText == "" {
		return nil, fmt.Errorf("CUSTOM_TEXT environment variable not set")
	}

	tzName := os.Getenv("TZ")
	if tzName == "" {
		tzName = "Asia/Tokyo"
		slog.Info(fmt.Sprintf("TZ environment variable not set, defaulting to %s", tzName))
	}

	location, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("error loading timezone %s: %w", tzName, err)
	}

	parsedTime, err := time.ParseInLocation(dateFormat, specificDateStr, location)
	if err != nil {
		return nil, fmt.Errorf("error parsing SPECIFIC_DATE '%s': %w. Expected format: %s", specificDateStr, err, dateFormat)
	}
	specificDate := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, location)

	return &config{
		specificDate:  specificDate,
		mkToken:       mkToken,
		misskeyAPIURL: buildAPIURL(misskeyHost),
		customText:    customText,
		location:      location,
	}, nil
}

func buildNoteRequest(apiURL, token, text string) (*http.Request, error) {
	payload := MisskeyNotePayload{
		Token: token,
		Text:  text,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := loadConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	nowInLocation := time.Now().In(cfg.location)
	daysPassed := calculateDaysPassed(cfg.specificDate, nowInLocation)

	if daysPassed == 0 && nowInLocation.Before(cfg.specificDate) {
		slog.Warn(fmt.Sprintf("Specific date is in the future. Posting 0."))
	}

	text := formatNoteText(cfg.customText, daysPassed)

	req, err := buildNoteRequest(cfg.misskeyAPIURL, cfg.mkToken, text)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("Error sending request to Misskey API (%s): %v", cfg.misskeyAPIURL, err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn(fmt.Sprintf("Error reading response body: %v", err))
	}

	slog.Info(fmt.Sprintf("API Response Status: %s", resp.Status))
	if len(respBody) > 0 {
		slog.Info(fmt.Sprintf("API Response Body: %s", string(respBody)))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error(fmt.Sprintf("Misskey API returned non-successful status: %s. Body: %s", resp.Status, string(respBody)))
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Successfully posted to Misskey: %.0f days passed since %s.", daysPassed, cfg.specificDate.Format(dateFormat)))
}
