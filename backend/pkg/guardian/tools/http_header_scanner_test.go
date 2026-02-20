package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPHeaderScanner(t *testing.T) {
	scanner := NewHTTPHeaderScanner()

	t.Run("scanner is available", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Error("expected scanner to be available")
		}
	})

	t.Run("has correct name", func(t *testing.T) {
		if scanner.Name() != "http_header_scanner" {
			t.Errorf("expected name to be http_header_scanner, got %s", scanner.Name())
		}
	})

	t.Run("has description", func(t *testing.T) {
		desc := scanner.Description()
		if desc == "" {
			t.Error("expected description to be non-empty")
		}
	})
}

func TestHTTPHeaderScannerExecute(t *testing.T) {
	scanner := NewHTTPHeaderScanner()
	ctx := context.Background()

	t.Run("fails with empty URL", func(t *testing.T) {
		args := HTTPHeaderScannerArgs{
			URL: "",
		}
		argsJSON, _ := json.Marshal(args)

		_, err := scanner.Execute(ctx, argsJSON)
		if err == nil {
			t.Error("expected error with empty URL")
		}
	})

	t.Run("scans server with no security headers", func(t *testing.T) {
		// Create test server without security headers
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer ts.Close()

		args := HTTPHeaderScannerArgs{
			URL: ts.URL,
		}
		argsJSON, _ := json.Marshal(args)

		result, err := scanner.Execute(ctx, argsJSON)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !result["success"].(bool) {
			t.Error("expected success to be true")
		}

		missingHeaders := result["missing_headers"].([]string)
		if len(missingHeaders) == 0 {
			t.Error("expected some missing headers")
		}

		securityScore := result["security_score"].(int)
		if securityScore != 0 {
			t.Errorf("expected security score to be 0, got %d", securityScore)
		}
	})

	t.Run("scans server with all security headers", func(t *testing.T) {
		// Create test server with all security headers
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Permissions-Policy", "geolocation=()")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer ts.Close()

		args := HTTPHeaderScannerArgs{
			URL: ts.URL,
		}
		argsJSON, _ := json.Marshal(args)

		result, err := scanner.Execute(ctx, argsJSON)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		missingHeaders := result["missing_headers"].([]string)
		if len(missingHeaders) != 0 {
			t.Errorf("expected no missing headers, got %d", len(missingHeaders))
		}

		securityScore := result["security_score"].(int)
		if securityScore != 100 {
			t.Errorf("expected security score to be 100, got %d", securityScore)
		}

		presentHeaders := result["present_headers"].(map[string]string)
		if len(presentHeaders) != 7 {
			t.Errorf("expected 7 present headers, got %d", len(presentHeaders))
		}
	})

	t.Run("scans server with partial security headers", func(t *testing.T) {
		// Create test server with some security headers
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer ts.Close()

		args := HTTPHeaderScannerArgs{
			URL: ts.URL,
		}
		argsJSON, _ := json.Marshal(args)

		result, err := scanner.Execute(ctx, argsJSON)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		missingHeaders := result["missing_headers"].([]string)
		if len(missingHeaders) != 5 {
			t.Errorf("expected 5 missing headers, got %d", len(missingHeaders))
		}

		presentHeaders := result["present_headers"].(map[string]string)
		if len(presentHeaders) != 2 {
			t.Errorf("expected 2 present headers, got %d", len(presentHeaders))
		}

		securityScore := result["security_score"].(int)
		expectedScore := (2 * 100) / 7 // 28
		if securityScore != expectedScore {
			t.Errorf("expected security score to be %d, got %d", expectedScore, securityScore)
		}
	})

	t.Run("fails with invalid URL", func(t *testing.T) {
		args := HTTPHeaderScannerArgs{
			URL: "http://this-domain-does-not-exist-12345.com",
		}
		argsJSON, _ := json.Marshal(args)

		_, err := scanner.Execute(ctx, argsJSON)
		if err == nil {
			t.Error("expected error with invalid URL")
		}
	})
}

func TestAnalyzeHeaders(t *testing.T) {
	scanner := NewHTTPHeaderScanner()

	t.Run("analyzes empty headers", func(t *testing.T) {
		headers := http.Header{}
		result := scanner.analyzeHeaders(headers)

		if len(result.MissingHeaders) != len(SecurityHeaders) {
			t.Errorf("expected %d missing headers, got %d", len(SecurityHeaders), len(result.MissingHeaders))
		}

		if len(result.PresentHeaders) != 0 {
			t.Errorf("expected 0 present headers, got %d", len(result.PresentHeaders))
		}

		if result.SecurityScore != 0 {
			t.Errorf("expected security score to be 0, got %d", result.SecurityScore)
		}
	})

	t.Run("analyzes complete headers", func(t *testing.T) {
		headers := http.Header{
			"Content-Security-Policy":   []string{"default-src 'self'"},
			"Strict-Transport-Security": []string{"max-age=31536000"},
			"X-Frame-Options":           []string{"DENY"},
			"X-Content-Type-Options":    []string{"nosniff"},
			"X-Xss-Protection":          []string{"1; mode=block"},
			"Referrer-Policy":           []string{"no-referrer"},
			"Permissions-Policy":        []string{"geolocation=()"},
		}
		result := scanner.analyzeHeaders(headers)

		if len(result.MissingHeaders) != 0 {
			t.Errorf("expected 0 missing headers, got %d", len(result.MissingHeaders))
		}

		if len(result.PresentHeaders) != len(SecurityHeaders) {
			t.Errorf("expected %d present headers, got %d", len(SecurityHeaders), len(result.PresentHeaders))
		}

		if result.SecurityScore != 100 {
			t.Errorf("expected security score to be 100, got %d", result.SecurityScore)
		}
	})
}

func TestGetRecommendation(t *testing.T) {
	scanner := NewHTTPHeaderScanner()

	t.Run("returns recommendation for known header", func(t *testing.T) {
		rec := scanner.getRecommendation("Content-Security-Policy")
		if rec == "" {
			t.Error("expected non-empty recommendation")
		}
		if rec != "Añadir: Content-Security-Policy: default-src 'self'" {
			t.Errorf("unexpected recommendation: %s", rec)
		}
	})

	t.Run("returns generic recommendation for unknown header", func(t *testing.T) {
		rec := scanner.getRecommendation("Unknown-Header")
		if rec == "" {
			t.Error("expected non-empty recommendation")
		}
	})
}
