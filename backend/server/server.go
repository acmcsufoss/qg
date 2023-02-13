package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler is the main entrypoint for the server.
type Handler struct {
	chi.Router
}

// NewHandler creates a new Handler.
func NewHandler() http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v0", func(r chi.Router) {
	})
}
