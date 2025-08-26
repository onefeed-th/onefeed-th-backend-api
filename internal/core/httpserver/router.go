package httpserver

import "net/http"

type Router struct {
	mux *http.ServeMux
}

func NewRouter(mux *http.ServeMux) *Router {
	return &Router{
		mux: mux,
	}
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.mux.Handle("GET "+path, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.mux.Handle("POST "+path, handler)
}
