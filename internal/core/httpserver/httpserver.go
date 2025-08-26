package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
)

type Endpoint[TReq any, TResp any] func() (fn Service[TReq, TResp])

func NewEndpoint[TReq any, TResp any](fn Service[TReq, TResp]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req TReq
		var finalRes dto.Response

		if r.Body != nil && r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				finalRes.Error = err.Error()
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(finalRes)
				return
			}
		}

		resp, err := fn(ctx, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			finalRes.Error = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}

		finalRes.Data = resp
		json.NewEncoder(w).Encode(finalRes)
	}
}
