package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"be-eventgate/internal/handlers"
	appmw "be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
	"be-eventgate/pkg/utils/response"
)

// New melakukan inisialisasi dan pengelompokan rute HTTP untuk fungsionalitas
// autentikasi serta implementasi sistem hierarki keamanan (Middleware).
func New(db *gorm.DB, jwtSecret string, jwtExpiryHrs int) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	authHandler := handlers.NewAuthHandler(db, jwtSecret, jwtExpiryHrs)
	userHandler := handlers.NewUserHandler(db)
	eventHandler := handlers.NewEventHandler(db)

	registerAPIRoutes := func(r chi.Router) {
		// Rute Publik (Tanpa Autentikasi)
		r.Post("/auth/login", authHandler.Login)

		// Rute Terproteksi (Memerlukan JWT Token)
		r.Group(func(r chi.Router) {
			r.Use(appmw.RequireAuth(jwtSecret))

			r.Get("/auth/me", userHandler.Me)

			// Rute Event Read (List & Detail)
			r.Get("/events", eventHandler.List)
			r.Get("/events/{id}", eventHandler.GetByID)

			// Area Proteksi RBAC: Manajemen Event (Create, Update, Delete)
			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole(models.RoleSuperAdmin, models.RoleAdminPanitia))
				r.Post("/events", eventHandler.Create)
				r.Put("/events/{id}", eventHandler.Update)
				r.Delete("/events/{id}", eventHandler.Delete)
			})

			// Area Proteksi RBAC: Memerlukan kewenangan peran super_admin
			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole(models.RoleSuperAdmin))
				r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
					response.Success(w, http.StatusOK, "pong, you are super_admin", nil)
				})
			})
		})
	}

	r.Route("/api", registerAPIRoutes)
	r.Route("/api/v1", registerAPIRoutes)

	return r
}
