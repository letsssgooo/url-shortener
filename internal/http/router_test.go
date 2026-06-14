package http

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ozon-intern/internal/service"
	"ozon-intern/internal/storage/memory"
)

type testGenerator struct {
	codes []string
}

func (g *testGenerator) Generate(ctx context.Context) (string, error) {
	if len(g.codes) == 0 {
		return "", errors.New("no codes")
	}

	code := g.codes[0]
	g.codes = g.codes[1:]

	return code, nil
}

func newTestRouter(t *testing.T) stdhttp.Handler {
	t.Helper()

	shortener := service.NewShortener(
		memory.NewStorage(),
		&testGenerator{codes: []string{"abcDEF123_"}},
	)

	return NewRouter(shortener, "https://short.example.com")
}

func newTestRouterWithGenerator(t *testing.T, generator *testGenerator) stdhttp.Handler {
	t.Helper()

	shortener := service.NewShortener(memory.NewStorage(), generator)

	return NewRouter(shortener, "https://short.example.com")
}

func TestRouterCreateLink(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/page"}`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, stdhttp.StatusCreated, response.Body.String())
	}

	var payload struct {
		ShortCode string `json:"short_code"`
		ShortURL  string `json:"short_url"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if payload.ShortCode != "abcDEF123_" {
		t.Fatalf("short_code: %s, want: %s", payload.ShortCode, "abcDEF123_")
	}
	if payload.ShortURL != "https://short.example.com/abcDEF123_" {
		t.Fatalf("short_url: %s, want: %s", payload.ShortURL, "https://short.example.com/abcDEF123_")
	}
}

func TestRouterCreateLinkInvalidJSON(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusBadRequest)
	}
}

func TestRouterCreateLinkInvalidURL(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"not-url"}`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusBadRequest)
	}
}

func TestRouterCreateLinkSameURLReturnsSameShortCode(t *testing.T) {
	router := newTestRouterWithGenerator(t, &testGenerator{
		codes: []string{"abcDEF123_", "unusedCode"},
	})

	firstRequest := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/page"}`))
	firstResponse := httptest.NewRecorder()
	router.ServeHTTP(firstResponse, firstRequest)

	secondRequest := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/page"}`))
	secondResponse := httptest.NewRecorder()
	router.ServeHTTP(secondResponse, secondRequest)

	if firstResponse.Code != stdhttp.StatusCreated {
		t.Fatalf("first status = %d, want %d", firstResponse.Code, stdhttp.StatusCreated)
	}
	if secondResponse.Code != stdhttp.StatusCreated {
		t.Fatalf("second status = %d, want %d", secondResponse.Code, stdhttp.StatusCreated)
	}

	var firstPayload struct {
		ShortCode string `json:"short_code"`
	}
	if err := json.NewDecoder(firstResponse.Body).Decode(&firstPayload); err != nil {
		t.Fatalf("Decode() first error = %v", err)
	}

	var secondPayload struct {
		ShortCode string `json:"short_code"`
	}
	if err := json.NewDecoder(secondResponse.Body).Decode(&secondPayload); err != nil {
		t.Fatalf("Decode() second error = %v", err)
	}

	if secondPayload.ShortCode != firstPayload.ShortCode {
		t.Fatalf("second short_code: %s, want: %s", secondPayload.ShortCode, firstPayload.ShortCode)
	}
}

func TestRouterGetLink(t *testing.T) {
	router := newTestRouter(t)
	createRequest := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/page"}`))
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)

	request := httptest.NewRequest(stdhttp.MethodGet, "/links/abcDEF123_", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, stdhttp.StatusOK, response.Body.String())
	}

	var payload struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if payload.URL != "https://example.com/page" {
		t.Fatalf("url: %s, want: %s", payload.URL, "https://example.com/page")
	}
}

func TestRouterGetLinkNotFound(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(stdhttp.MethodGet, "/links/missing000", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusNotFound)
	}
}

func TestRouterGetLinkInvalidShortCode(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(stdhttp.MethodGet, "/links/invalid", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusBadRequest)
	}
}

func TestRouterRedirect(t *testing.T) {
	router := newTestRouter(t)
	createRequest := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/page"}`))
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)

	request := httptest.NewRequest(stdhttp.MethodGet, "/abcDEF123_", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusFound {
		t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusFound)
	}
	if location := response.Header().Get("Location"); location != "https://example.com/page" {
		t.Fatalf("Location: %s, want: %s", location, "https://example.com/page")
	}
}
