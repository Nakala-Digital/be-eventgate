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
	questionHandler := handlers.NewQuestionHandler(db)
	registrationHandler := handlers.NewRegistrationHandler(db)

	registerAPIRoutes := func(r chi.Router) {
		// Rute Publik (Tanpa Autentikasi)
		r.Post("/auth/login", authHandler.Login)

		// Implementasi EVG-49: Public Participant Registration API.
		// Rute di bawah subtree `/api/public` disediakan khusus untuk entitas publik (tanpa akun).
		// Area ini sengaja mengecualikan middleware `RequireAuth` agar dapat diakses tanpa token.
		r.Route("/public", func(r chi.Router) {
			r.Get("/events/{id}/ticket-types", registrationHandler.ListPublicTicketTypes)
			r.Get("/events/{id}/questions", registrationHandler.ListPublicQuestions)
			r.Post("/events/{id}/register", registrationHandler.Register)
			r.Get("/registrations/{code}", registrationHandler.GetByCode)
		})

		// Rute Terproteksi (Memerlukan JWT Token)
		r.Group(func(r chi.Router) {
			r.Use(appmw.RequireAuth(jwtSecret))

			r.Get("/auth/me", userHandler.Me)

			// Rute Event Read (List & Detail)
			r.Get("/events", eventHandler.List)
			r.Get("/events/{id}", eventHandler.GetByID)
			r.Get("/events/{id}/approval-logs", eventHandler.ListApprovalLogs)

			// Khusus admin_panitia: submit approval.
			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole(models.RoleAdminPanitia))
				r.Post("/events/{id}/submit", eventHandler.SubmitForApproval)
			})

			// Khusus super_admin/school_reviewer: approve/reject/revisi.
			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole(models.RoleSuperAdmin, models.RoleSchoolReviewer))
				r.Post("/events/{id}/approve", eventHandler.ApproveEvent)
				r.Post("/events/{id}/reject", eventHandler.RejectEvent)
				r.Post("/events/{id}/request-revision", eventHandler.RequestRevision)
			})

			// Khusus super_admin: publish/unpublish.
			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireRole(models.RoleSuperAdmin))
				r.Post("/events/{id}/publish", eventHandler.PublishEvent)
				r.Post("/events/{id}/unpublish", eventHandler.UnpublishEvent)
			})

			// Area EVG-47: Skema Formulir Dinamis
			r.Route("/events/{id}/questions", func(r chi.Router) {
				// Akses baca: Pemilik (admin_panitia), super_admin, atau school_reviewer.
				// Mengikuti aturan visibilitas kegiatan yang sama dengan endpoint GetByID.
				r.Get("/", questionHandler.ListByEvent)

				// Akses pengelolaan (buat/ubah/hapus): Pemilik kegiatan atau super_admin.
				// Validasi hak akses dilakukan di dalam handler berdasarkan data kegiatan.
				r.Post("/", questionHandler.Create)
				r.Put("/{questionID}", questionHandler.Update)
				r.Delete("/{questionID}", questionHandler.Delete)
			})

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
