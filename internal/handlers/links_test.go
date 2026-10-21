package handlers

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepository struct {
	links map[string]string
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		links: make(map[string]string),
	}
}

func (f *fakeRepository) CreateLink(
	ctx context.Context,
	targetURL string,
) (string, error) {
	code := "test1234"
	f.links[code] = targetURL
	return code, nil
}

func (f *fakeRepository) GetLink(
	ctx context.Context,
	code string,
) (string, error) {
	target, ok := f.links[code]
	if !ok {
		return "", pgx.ErrNoRows
	}
	return target, nil
}

func TestShortenAndRedirect(t *testing.T) {
	repo := newFakeRepository()
	handler := NewLinkHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", handler.Shorten)
	mux.HandleFunc("GET /{code}", handler.Redirect)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(`{"url":"https://go.dev/docs"}`),
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Code      string `json:"code"`
		TargetURL string `json:"target_url"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Code == "" {
		t.Fatal("expected a short code")
	}

	redirectReq := httptest.NewRequest(
		http.MethodGet,
		"/"+response.Code,
		nil,
	)
	redirectRec := httptest.NewRecorder()
	mux.ServeHTTP(redirectRec, redirectReq)

	if redirectRec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", redirectRec.Code)
	}

	if got := redirectRec.Header().Get("Location"); got != response.TargetURL {
		t.Fatalf("expected Location %q, got %q", response.TargetURL, got)
	}
}

func TestShortenRejectsInvalidURL(t *testing.T) {
	handler := NewLinkHandler(newFakeRepository())

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(`{"url":"javascript:alert(1)"}`),
	)
	rec := httptest.NewRecorder()

	handler.Shorten(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRedirectUnknownCode(t *testing.T) {
	handler := NewLinkHandler(newFakeRepository())

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	req.SetPathValue("code", "unknown")
	rec := httptest.NewRecorder()

	handler.Redirect(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
