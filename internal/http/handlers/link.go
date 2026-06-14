package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"ozon-intern/internal/generator"
	"ozon-intern/internal/service"
)

type LinkHandler struct {
	shortener *service.Shortener
	baseURL   string
}

func NewLinkHandler(shortener *service.Shortener, baseURL string) *LinkHandler {
	return &LinkHandler{
		shortener: shortener,
		baseURL:   strings.TrimRight(baseURL, "/"),
	}
}

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	if h.shortener == nil {
		writeError(w, http.StatusInternalServerError, "service is not configured")
		return
	}

	var request createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	originalURL := strings.TrimSpace(request.URL)
	if !isValidOriginalURL(originalURL) {
		writeError(w, http.StatusBadRequest, "invalid url")
		return
	}

	shortCode, err := h.shortener.Shorten(r.Context(), originalURL)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createLinkResponse{
		ShortCode: shortCode,
		ShortURL:  buildShortURL(h.baseURL, shortCode),
	})
}

func (h *LinkHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.shortener == nil {
		writeError(w, http.StatusInternalServerError, "service is not configured")
		return
	}

	shortCode := chi.URLParam(r, "shortCode")
	if !isValidShortCode(shortCode) {
		writeError(w, http.StatusBadRequest, "invalid short code")
		return
	}

	originalURL, err := h.shortener.Resolve(r.Context(), shortCode)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, getLinkResponse{URL: originalURL})
}

func isValidOriginalURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	return (parsedURL.Scheme == "http" || parsedURL.Scheme == "https") && parsedURL.Host != ""
}

func isValidShortCode(shortCode string) bool {
	if len(shortCode) != generator.CodeLength {
		return false
	}

	for _, char := range shortCode {
		if !strings.ContainsRune(generator.Alphabet, char) {
			return false
		}
	}

	return true
}

func buildShortURL(baseURL string, shortCode string) string {
	return baseURL + "/" + shortCode
}
