package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/onefeed-th/onefeed-th-backend-api/config"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/rds"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/db"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/logger"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/middleware"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/repository"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/routes"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/service"
)

func main() {
	log := logger.New("application-go")
	// setup signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// initialize configuration
	if err := config.Init(ctx, "config/config.yaml"); err != nil {
		log.Error(ctx, "Failed to initialize configuration", "error", err)
		return
	}
	cfg := config.GetConfig()

	// initialize database
	if err := db.InitDB(); err != nil {
		log.Error(ctx, "Failed to initialize database", "error", err)
	}

	// initialize Redis
	if err := rds.InitRedis(ctx); err != nil {
		log.Error(ctx, "Failed to initialize Redis", "error", err)
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
		log.Info(ctx, "Starting REST Server", "port", cfg.RestServer.Port)
		log.Info(ctx, "Local server", "url", fmt.Sprintf("http://localhost:%d", cfg.RestServer.Port))
		log.Info(ctx, "waiting for request...")

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(ctx, "Failed to serve", "error", err)
			stop()
		}
	}()

	// wait for the context to be canceled (i.e., SIGINT or SIGTERM)
	<-ctx.Done()
	log.Info(ctx, "Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error(ctx, "Server shutdown failed", "error", err)
	}

	// Close database connections
	db.CloseDB()
	log.Info(ctx, "Database connections closed")

	// Close Redis connections
	if err := rds.CloseRedis(); err != nil {
		log.Error(ctx, "Redis shutdown failed", "error", err)
	} else {
		log.Info(ctx, "Redis connections closed")
	}

	// Sync logger to flush any buffered entries
	if err := logger.Sync(); err != nil {
		log.Error(ctx, "Logger sync failed", "error", err)
	}

	log.Info(ctx, "Server gracefully stopped")
}
