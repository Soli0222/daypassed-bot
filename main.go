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

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil))) // Ensure logs go to stdout for container logging

	specificDateStr := os.Getenv("SPECIFIC_DATE")
	if specificDateStr == "" {
		slog.Error("SPECIFIC_DATE environment variable not set. Expected format: YYYY-MM-DD")
		os.Exit(1)
	}

	mkToken := os.Getenv("MK_TOKEN")
	if mkToken == "" {
		slog.Error("MK_TOKEN environment variable not set.")
		os.Exit(1)
	}

	misskeyHost := os.Getenv("MISSKEY_HOST")
	if misskeyHost == "" {
		slog.Error("MISSKEY_HOST environment variable not set.")
		os.Exit(1)
	}
	// Check if the host is localhost, then use http instead of https
	scheme := "https"
	if strings.Contains(misskeyHost, "localhost") {
		scheme = "http"
	}
	misskeyAPIURL := fmt.Sprintf("%s://%s/api/notes/create", scheme, misskeyHost)

	tzName := os.Getenv("TZ")
	if tzName == "" {
		tzName = "Asia/Tokyo" // Default timezone
		slog.Info(fmt.Sprintf("TZ environment variable not set, defaulting to %s", tzName))
	}

	location, err := time.LoadLocation(tzName)
	if err != nil {
		slog.Error(fmt.Sprintf("Error loading timezone %s: %v", tzName, err))
		os.Exit(1)
	}

	customText := os.Getenv("CUSTOM_TEXT")
	if customText == "" {
		slog.Error("CUSTOM_TEXT environment variable not set.")
		os.Exit(1)
	}

	// Parse specificDateStr as a date in the specified timezone.
	// This represents the start of the day for the specific date.
	parsedTime, err := time.ParseInLocation(dateFormat, specificDateStr, location)
	if err != nil {
		slog.Error(fmt.Sprintf("Error parsing SPECIFIC_DATE '%s': %v. Expected format: %s", specificDateStr, err, dateFormat))
		os.Exit(1)
	}
	specificDateStartOfDay := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, location)

	nowInLocation := time.Now().In(location)

	// Calculate days passed
	timeDifference := nowInLocation.Sub(specificDateStartOfDay)
	daysPassed := math.Floor(timeDifference.Hours() / 24)

	if daysPassed < 0 {
		slog.Warn(fmt.Sprintf("Specific date %s is in the future. Days passed is negative (%.0f). Posting 0.", specificDateStr, daysPassed))
		daysPassed = 0 // Or handle as an error, depending on desired behavior
	}

	text := fmt.Sprintf("<center>%s\n\n$[jelly $[sparkle %.0f日]]</center>", customText, daysPassed)

	payload := MisskeyNotePayload{
		Token: mkToken,
		Text:  text,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error(fmt.Sprintf("Error marshalling JSON payload: %v", err))
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", misskeyAPIURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating request: %v", err))
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second} // Increased timeout
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("Error sending request to Misskey API (%s): %v", misskeyAPIURL, err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// Log error but continue to log status, as status might be informative
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

	slog.Info(fmt.Sprintf("Successfully posted to Misskey: %s days passed since %s.", fmt.Sprintf("%.0f", daysPassed), specificDateStr))
}
