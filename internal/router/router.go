package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"be-eventgate/internal/handlers"
	appmw "be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
)

// New melakukan inisialisasi dan pengelompokan rute HTTP untuk fungsionalitas
// autentikasi serta implementasi sistem hierarki keamanan (Middleware).
func New(db *gorm.DB, jwtSecret string, jwtExpiryHrs int) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	authHandler := handlers.NewAuthHandler(db, jwtSecret, jwtExpiryHrs)
	userHandler := handlers.NewUserHandler(db)

	r.Route("/api", func(r chi.Router) {
		// Rute Publik (Tanpa Autentikasi)
		r.Post("/auth/login", authHandler.Login)

		// Rute Terproteksi (Memerlukan JWT Token)
		r.Group(func(r chi.Router) {
			r.Use(appmw.RequireAuth(jwtSecret))

			r.Get("/auth/me", userHandler.Me)

			// Area Proteksi RBAC: Memerlukan kewenangan peran super_admin
			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole(models.RoleSuperAdmin))
				r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"message":"pong, you are super_admin"}`))
				})
			})
		})
	})

	return r
}
