package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"ozon-intern/internal/config"
	"ozon-intern/internal/generator"
	httpRoute "ozon-intern/internal/http"
	"ozon-intern/internal/service"
	"ozon-intern/internal/storage"
	"ozon-intern/internal/storage/memory"
	"ozon-intern/internal/storage/postgres"
)

const shutdownTimeout = 5 * time.Second

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	store, cleanup, err := newStorage(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	codeGenerator := generator.NewRandomGenerator()
	shortener := service.NewShortener(store, codeGenerator)
	router := httpRoute.NewRouter(shortener, cfg.BaseURL)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return runServer(server)
}

func newStorage(ctx context.Context, cfg config.Config) (storage.Storage, func(), error) {
	switch cfg.StorageType {
	case config.StorageTypeMemory:
		return memory.NewStorage(), func() {}, nil
	case config.StorageTypePostgres:
		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, func() {}, fmt.Errorf("create postgres pool: %w", err)
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, func() {}, fmt.Errorf("ping postgres: %w", err)
		}

		return postgres.NewStorage(pool), pool.Close, nil
	default:
		return nil, func() {}, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
}

func runServer(server *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("http server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	slog.Info("http server shutdown started")
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	if err := <-errCh; err != nil {
		return err
	}

	slog.Info("http server stopped")
	return nil
}
