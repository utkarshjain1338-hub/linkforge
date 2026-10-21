package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() http.Handler {
	store := NewLinkStore()
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", store.Shorten)
	mux.HandleFunc("GET /{code}", store.Redirect)

	return mux
}

func TestShortenAndRedirect(t *testing.T) {
	server := newTestServer()

	body := `{"url":"https://go.dev/docs"}`
	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s",
			http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response struct {
		Code      string `json:"code"`
		ShortURL  string `json:"short_url"`
		TargetURL string `json:"target_url"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if response.Code == "" {
		t.Fatal("expected a non-empty short code")
	}

	if response.TargetURL != "https://go.dev/docs" {
		t.Fatalf("unexpected target URL: %s", response.TargetURL)
	}

	redirectReq := httptest.NewRequest(
		http.MethodGet,
		"/"+response.Code,
		nil,
	)
	redirectRec := httptest.NewRecorder()

	server.ServeHTTP(redirectRec, redirectReq)

	if redirectRec.Code != http.StatusFound {
		t.Fatalf("expected redirect status %d, got %d",
			http.StatusFound, redirectRec.Code)
	}

	if got := redirectRec.Header().Get("Location"); got != response.TargetURL {
		t.Fatalf("expected Location %q, got %q", response.TargetURL, got)
	}
}

func TestShortenRejectsInvalidURLs(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid JSON",
			body: `{"url":`,
		},
		{
			name: "unsupported scheme",
			body: `{"url":"javascript:alert(1)"}`,
		},
		{
			name: "missing host",
			body: `{"url":"https://"}`,
		},
		{
			name: "empty URL",
			body: `{"url":""}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer()
			req := httptest.NewRequest(
				http.MethodPost,
				"/shorten",
				strings.NewReader(tc.body),
			)
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d",
					http.StatusBadRequest, rec.Code)
			}
		})
	}
}

func TestRedirectUnknownCode(t *testing.T) {
	server := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/unknown-code", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d",
			http.StatusNotFound, rec.Code)
	}
}
