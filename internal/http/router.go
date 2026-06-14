package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"ozon-intern/internal/http/handlers"
	"ozon-intern/internal/service"
)

func NewRouter(shortener *service.Shortener, baseURL string) http.Handler {
	router := chi.NewRouter()

	linkHandler := handlers.NewLinkHandler(shortener, baseURL)
	redirectHandler := handlers.NewRedirectHandler(shortener)

	router.Post("/links", linkHandler.Create)
	router.Get("/links/{shortCode}", linkHandler.Get)
	router.Get("/{shortCode}", redirectHandler.Redirect)

	return router
}
