package main

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

	"github.com/jackc/pgx/v5"
	repo "github.com/nikallow/bookstores-api/internal/adapters/postgres/sqlc"
	"github.com/nikallow/bookstores-api/internal/books"
	"github.com/nikallow/bookstores-api/internal/config"
	"github.com/nikallow/bookstores-api/internal/inventory"
	"github.com/nikallow/bookstores-api/internal/logger"
	"github.com/nikallow/bookstores-api/internal/stores"
)

func main() {
	// Config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = ""
		slog.Warn("Config file is not found, loading from env")
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Logger
	l := logger.SetupLogger(cfg)
	l.Info("Logger initialized", "env", cfg.Env)

	// PostgreSQL
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	conn, err := pgx.Connect(dbCtx, cfg.Database.GetDSN())
	if err != nil {
		l.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	l.Info("Connected to database", "db_name", cfg.Database.DBName, "host", cfg.Database.Host)

	// DI
	dbQuerier := repo.New(conn)

	// Services and Handlers
	storeService := stores.NewService(dbQuerier)
	storeHandler := stores.NewHandler(storeService)

	booksService := books.NewService(dbQuerier)
	booksHandler := books.NewHandler(booksService)

	inventoryService := inventory.NewService(dbQuerier, conn)
	inventoryHandler := inventory.NewHandler(inventoryService)

	apiDeps := &APIDependencies{
		Logger:           l,
		StoreHandler:     storeHandler,
		BooksHandler:     booksHandler,
		InventoryHandler: inventoryHandler,
	}

	// Launch HTTP server
	httpServer := NewHTTPServer(cfg, apiDeps)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()
	l.Info("HTTP server started", "addr", httpServer.Addr)

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		l.Error("HTTP server shutdown failed", "error", err)
	} else {
		l.Info("Server gracefully stopped")
	}
}

func NewHTTPServer(cfg *config.Config, deps *APIDependencies) *http.Server {
	router := MountAPI(deps)
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Service.Host, cfg.Service.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
