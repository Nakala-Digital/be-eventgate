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

// New menyusun semua route EVG-41 (auth & RBAC).
//
// SESUAIKAN: kalau project Anda sudah punya router.go / main.go sendiri,
// pindahkan isi function ini ke dalam struktur yang sudah ada. Yang penting
// urutan middleware-nya: RequireAuth dulu, baru RequireRole.
func New(db *gorm.DB, jwtSecret string, jwtExpiryHrs int) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	authHandler := handlers.NewAuthHandler(db, jwtSecret, jwtExpiryHrs)
	userHandler := handlers.NewUserHandler(db)

	r.Route("/api", func(r chi.Router) {
		// Public
		r.Post("/auth/login", authHandler.Login)

		// Protected (butuh token valid)
		r.Group(func(r chi.Router) {
			r.Use(appmw.RequireAuth(jwtSecret))

			r.Get("/auth/me", userHandler.Me)

			// Contoh route yang dibatasi role tertentu (pola ini dipakai untuk
			// endpoint-endpoint task berikutnya, mis. EVG-45/47/49).
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
