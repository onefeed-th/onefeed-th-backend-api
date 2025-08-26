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

	// initialize database
	if err := db.InitDB(); err != nil {
		log.Println("Failed to initialize database:", err)
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
		Addr:    fmt.Sprintf(":%d", 3000),
		Handler: httpHandler,
	}

	go func() {
		log.Printf("Starting REST Server on port %d\n", 3000)
		log.Printf("Local : http://localhost:%d\n", 3000)
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

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v\n", err)
	}

	log.Println("Server gracefully stopped")
}
