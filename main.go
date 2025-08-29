package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	// setup signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// initialize configuration
	if err := config.Init(ctx, "config/config.yaml"); err != nil {
		log.Printf("Failed to initialize configuration: %v", err)
		return
	}
	cfg := config.GetConfig()

	// initialize database
	if err := db.InitDB(); err != nil {
		log.Println("Failed to initialize database:", err)
	}

	// initialize Redis
	if err := rds.InitRedis(ctx); err != nil {
		log.Println("Failed to initialize Redis:", err)
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
		log.Printf("Starting REST Server on port %d\n", cfg.RestServer.Port)
		log.Printf("Local : http://localhost:%d\n", cfg.RestServer.Port)
		log.Println("waiting for request...")

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("failed to served %s\n", err)
			stop()
		}
	}()

	// wait for the context to be canceled (i.e., SIGINT or SIGTERM)
	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v\n", err)
	}

	// Close database connections
	db.CloseDB()
	log.Println("Database connections closed")

	// Close Redis connections
	if err := rds.CloseRedis(); err != nil {
		log.Printf("Redis shutdown failed: %v\n", err)
	} else {
		log.Println("Redis connections closed")
	}

	// Sync logger to flush any buffered entries
	if err := logger.Sync(); err != nil {
		log.Printf("Logger sync failed: %v\n", err)
	}

	log.Println("Server gracefully stopped")
}
