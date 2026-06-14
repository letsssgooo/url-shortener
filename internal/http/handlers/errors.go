package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"ozon-intern/internal/service"
	"ozon-intern/internal/storage"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal json response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		writeError(w, http.StatusNotFound, "link not found")
	case errors.Is(err, service.ErrTooManyCollisions):
		writeError(w, http.StatusConflict, "could not generate unique short code")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
