package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type LinkStore struct {
	mu    sync.RWMutex
	links map[string]string
}

func NewLinkStore() *LinkStore {
	return &LinkStore{
		links: make(map[string]string),
	}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"short_url"`
	TargetURL string `json:"target_url"`
}

func (s *LinkStore) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	parsed, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		http.Error(w, "a valid HTTP or HTTPS URL is required", http.StatusBadRequest)
		return
	}

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "could not generate code", http.StatusInternalServerError)
		return
	}
	code := hex.EncodeToString(bytes)

	s.mu.Lock()
	s.links[code] = req.URL
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{
		Code:      code,
		ShortURL:  "http://localhost:8080/" + code,
		TargetURL: req.URL,
	})
}

func (s *LinkStore) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	s.mu.RLock()
	target, ok := s.links[code]
	s.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, target, http.StatusFound)
}
