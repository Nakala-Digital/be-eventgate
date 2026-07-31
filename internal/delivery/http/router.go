package http

import (
    "net/http"

    customMiddleware "be-eventgate/internal/delivery/http/middleware"
    "be-eventgate/pkg/utils/response"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(customMiddleware.CORS)

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        response.Success(w, http.StatusOK, "Welcome to EventGate API", nil)
    })

    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
            response.Success(w, http.StatusOK, "API is running healthy", map[string]string{
                "status": "UP",
            })
        })
    })

    return r
}
