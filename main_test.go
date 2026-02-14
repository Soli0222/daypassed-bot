package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"
)

func TestBuildAPIURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "normal host uses https",
			host: "mi.example.com",
			want: "https://mi.example.com/api/notes/create",
		},
		{
			name: "localhost uses http",
			host: "localhost",
			want: "http://localhost/api/notes/create",
		},
		{
			name: "localhost with port uses http",
			host: "localhost:3000",
			want: "http://localhost:3000/api/notes/create",
		},
		{
			name: "subdomain containing localhost uses http",
			host: "api.localhost",
			want: "http://api.localhost/api/notes/create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAPIURL(tt.host)
			if got != tt.want {
				t.Errorf("buildAPIURL(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

func TestCalculateDaysPassed(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name         string
		specificDate time.Time
		now          time.Time
		want         float64
	}{
		{
			name:         "1 day ago",
			specificDate: time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
			now:          time.Date(2025, 1, 2, 12, 0, 0, 0, loc),
			want:         1,
		},
		{
			name:         "100 days ago",
			specificDate: time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
			now:          time.Date(2025, 4, 11, 0, 0, 0, 0, loc),
			want:         100,
		},
		{
			name:         "same day returns 0",
			specificDate: time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
			now:          time.Date(2025, 1, 1, 15, 0, 0, 0, loc),
			want:         0,
		},
		{
			name:         "future date returns 0",
			specificDate: time.Date(2025, 12, 31, 0, 0, 0, 0, loc),
			now:          time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
			want:         0,
		},
		{
			name:         "23 hours is 0 days",
			specificDate: time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
			now:          time.Date(2025, 1, 1, 23, 0, 0, 0, loc),
			want:         0,
		},
		{
			name:         "25 hours is 1 day",
			specificDate: time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
			now:          time.Date(2025, 1, 2, 1, 0, 0, 0, loc),
			want:         1,
		},
		{
			name:         "large number of days",
			specificDate: time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
			now:          time.Date(2025, 4, 1, 0, 0, 0, 0, loc),
			want:         90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDaysPassed(tt.specificDate, tt.now)
			if got != tt.want {
				t.Errorf("calculateDaysPassed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatNoteText(t *testing.T) {
	tests := []struct {
		name       string
		customText string
		daysPassed float64
		want       string
	}{
		{
			name:       "normal formatting",
			customText: "今日は起動してから",
			daysPassed: 100,
			want:       "<center>今日は起動してから\n\n$[jelly $[sparkle 100日]]</center>",
		},
		{
			name:       "zero days",
			customText: "テスト",
			daysPassed: 0,
			want:       "<center>テスト\n\n$[jelly $[sparkle 0日]]</center>",
		},
		{
			name:       "large number",
			customText: "経過日数",
			daysPassed: 9999,
			want:       "<center>経過日数\n\n$[jelly $[sparkle 9999日]]</center>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatNoteText(tt.customText, tt.daysPassed)
			if got != tt.want {
				t.Errorf("formatNoteText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Helper to set all required env vars
	setAllEnvs := func(t *testing.T) {
		t.Helper()
		t.Setenv("SPECIFIC_DATE", "2023-04-07")
		t.Setenv("MK_TOKEN", "test-token")
		t.Setenv("MISSKEY_HOST", "mi.example.com")
		t.Setenv("CUSTOM_TEXT", "テスト")
		t.Setenv("TZ", "Asia/Tokyo")
	}

	t.Run("all env vars set", func(t *testing.T) {
		setAllEnvs(t)

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.mkToken != "test-token" {
			t.Errorf("mkToken = %q, want %q", cfg.mkToken, "test-token")
		}
		if cfg.customText != "テスト" {
			t.Errorf("customText = %q, want %q", cfg.customText, "テスト")
		}
		if cfg.misskeyAPIURL != "https://mi.example.com/api/notes/create" {
			t.Errorf("misskeyAPIURL = %q, want %q", cfg.misskeyAPIURL, "https://mi.example.com/api/notes/create")
		}
		if cfg.location.String() != "Asia/Tokyo" {
			t.Errorf("location = %q, want %q", cfg.location.String(), "Asia/Tokyo")
		}
		wantDate := time.Date(2023, 4, 7, 0, 0, 0, 0, cfg.location)
		if !cfg.specificDate.Equal(wantDate) {
			t.Errorf("specificDate = %v, want %v", cfg.specificDate, wantDate)
		}
	})

	t.Run("SPECIFIC_DATE not set", func(t *testing.T) {
		setAllEnvs(t)
		os.Unsetenv("SPECIFIC_DATE")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("MK_TOKEN not set", func(t *testing.T) {
		setAllEnvs(t)
		os.Unsetenv("MK_TOKEN")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("MISSKEY_HOST not set", func(t *testing.T) {
		setAllEnvs(t)
		os.Unsetenv("MISSKEY_HOST")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("CUSTOM_TEXT not set", func(t *testing.T) {
		setAllEnvs(t)
		os.Unsetenv("CUSTOM_TEXT")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("TZ not set defaults to Asia/Tokyo", func(t *testing.T) {
		setAllEnvs(t)
		os.Unsetenv("TZ")

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.location.String() != "Asia/Tokyo" {
			t.Errorf("location = %q, want %q", cfg.location.String(), "Asia/Tokyo")
		}
	})

	t.Run("invalid date format", func(t *testing.T) {
		setAllEnvs(t)
		t.Setenv("SPECIFIC_DATE", "04-07-2023")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("expected error for invalid date format, got nil")
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		setAllEnvs(t)
		t.Setenv("TZ", "Invalid/Timezone")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("expected error for invalid timezone, got nil")
		}
	})

	t.Run("localhost host uses http", func(t *testing.T) {
		setAllEnvs(t)
		t.Setenv("MISSKEY_HOST", "localhost:3000")

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.misskeyAPIURL != "http://localhost:3000/api/notes/create" {
			t.Errorf("misskeyAPIURL = %q, want %q", cfg.misskeyAPIURL, "http://localhost:3000/api/notes/create")
		}
	})
}

func TestBuildNoteRequest(t *testing.T) {
	apiURL := "https://mi.example.com/api/notes/create"
	token := "test-token"
	text := "test note"

	t.Run("creates valid POST request", func(t *testing.T) {
		req, err := buildNoteRequest(apiURL, token, text)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if req.Method != "POST" {
			t.Errorf("method = %q, want %q", req.Method, "POST")
		}

		if req.URL.String() != apiURL {
			t.Errorf("url = %q, want %q", req.URL.String(), apiURL)
		}

		if ct := req.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
	})

	t.Run("body contains token and text", func(t *testing.T) {
		req, err := buildNoteRequest(apiURL, token, text)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("error reading body: %v", err)
		}

		var payload MisskeyNotePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("error unmarshalling body: %v", err)
		}

		if payload.Token != token {
			t.Errorf("token = %q, want %q", payload.Token, token)
		}
		if payload.Text != text {
			t.Errorf("text = %q, want %q", payload.Text, text)
		}
	})
}

func TestMisskeyNotePayloadJSON(t *testing.T) {
	payload := MisskeyNotePayload{
		Token: "my-token",
		Text:  "hello",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshalling: %v", err)
	}

	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("error unmarshalling: %v", err)
	}

	if m["i"] != "my-token" {
		t.Errorf("json field 'i' = %q, want %q", m["i"], "my-token")
	}
	if m["text"] != "hello" {
		t.Errorf("json field 'text' = %q, want %q", m["text"], "hello")
	}
	if len(m) != 2 {
		t.Errorf("expected 2 fields, got %d: %v", len(m), m)
	}
}
