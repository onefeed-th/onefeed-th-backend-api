package routes

import (
	"net/http"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/httpserver"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/service"
)

func RegisterRoutes(service service.Service) http.Handler {
	mux := http.NewServeMux()
	r := httpserver.NewRouter(mux)

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	// Server
	{
		r.Get("/health",
			httpserver.NewEndpoint(
				service.HealthCheck,
			),
		)
	}

	// collector
	{
		r.Get("/internal/collect",
			httpserver.NewEndpoint(
				service.CollectNewsFromSource,
			),
		)
	}

	return mux
}
