package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"ozon-intern/internal/service"
)

type RedirectHandler struct {
	shortener *service.Shortener
}

func NewRedirectHandler(shortener *service.Shortener) *RedirectHandler {
	return &RedirectHandler{shortener: shortener}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, originalURL, http.StatusFound)
}
