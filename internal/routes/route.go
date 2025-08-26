package routes

import (
	"net/http"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/service"
)

func RegisterRoutes(service *service.Service) http.Handler {
	mux := http.NewServeMux()

	return mux
}
