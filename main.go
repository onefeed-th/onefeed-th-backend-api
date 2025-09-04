package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/onefeed-th/onefeed-th-backend-api/config"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/rds"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/db"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/middleware"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/repository"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/routes"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/service"
)

func main() {
	// setup signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// initialize configuration
	if err := config.Init(ctx, "config/config.yaml"); err != nil {
		slog.Error("Failed to initialize configuration", "error", err)
		return
	}
	cfg := config.GetConfig()

	// initialize database
	if err := db.InitDB(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
	}

	// initialize Redis
	if err := rds.InitRedis(ctx); err != nil {
		slog.Error("Failed to initialize Redis", "error", err)
	}

	// initialize repository
	repo := repository.NewRepository()

	// initialize service
	service := service.NewService(repo)

	// initialize mux
	handler := routes.RegisterRoutes(service)
	handler = middleware.LogRequest(handler)
	handler = middleware.RecoverPanic(handler)

	// global middlewares
	var httpHandler http.Handler = handler

	// create configure http server
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.RestServer.Port),
		Handler: httpHandler,
	}

	go func() {
		slog.Info("Starting REST Server", "port", cfg.RestServer.Port)
		slog.Info("Local server", "url", fmt.Sprintf("http://localhost:%d", cfg.RestServer.Port))
		slog.Info("waiting for request...")

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to serve", "error", err)
			stop()
		}
	}()

	// wait for the context to be canceled (i.e., SIGINT or SIGTERM)
	<-ctx.Done()
	slog.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}

	// Close database connections
	db.CloseDB()
	slog.Info("Database connections closed")

	// Close Redis connections
	if err := rds.CloseRedis(); err != nil {
		slog.Error("Redis shutdown failed", "error", err)
	} else {
		slog.Info("Redis connections closed")
	}

	slog.Info("Server gracefully stopped")
}
