package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type LinkRepository interface {
	CreateLink(ctx context.Context, targetURL string) (string, error)
	GetLink(ctx context.Context, code string) (string, error)
}

type LinkHandler struct {
	db LinkRepository
}

func NewLinkHandler(db LinkRepository) *LinkHandler {
	return &LinkHandler{db: db}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"short_url"`
	TargetURL string `json:"target_url"`
}

func (h *LinkHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	parsed, err := url.ParseRequestURI(req.URL)
	if err != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Host == "" {
		http.Error(w, "a valid HTTP or HTTPS URL is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	code, err := h.db.CreateLink(ctx, req.URL)
	if err != nil {
		http.Error(w, "could not create short link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(shortenResponse{
		Code:      code,
		ShortURL:  "http://localhost:8080/" + code,
		TargetURL: req.URL,
	})
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	target, err := h.db.GetLink(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "could not retrieve link", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, target, http.StatusFound)
}
