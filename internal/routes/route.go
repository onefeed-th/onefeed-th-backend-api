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
		r.Post("/internal/collect",
			httpserver.NewEndpoint(
				service.CollectNewsFromSource,
			),
		)
		r.Post("/internal/delete-old-news",
			httpserver.NewEndpoint(
				service.RemoveOldNews,
			),
		)
	}

	// news
	{
		r.Post("/news",
			httpserver.NewEndpoint(
				service.GetNews,
			),
		)
	}

	// tags
	{
		r.Get("/tags",
			httpserver.NewEndpoint(
				service.GetAllTags,
			),
		)
	}

	// backoffice
	{
		r.Post("/backoffice/get-sources",
			httpserver.NewEndpoint(
				service.GetAllSourceByPagination,
			),
		)
		r.Post("/backoffice/create-source",
			httpserver.NewEndpoint(
				service.CreateSource,
			),
		)
	}

	return mux
}
