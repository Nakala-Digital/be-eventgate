package handlers

import (
	"net/http"

	"gorm.io/gorm"

	"be-eventgate/internal/httpx"
	"be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// UserProfile adalah representasi user yang aman ditampilkan ke client
// (tidak menyertakan password hash).
type UserProfile struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleName string `json:"role_name"`
	IsActive bool   `json:"is_active"`
}

func toUserProfile(u models.User) UserProfile {
	return UserProfile{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		RoleName: u.Role.RoleName,
		IsActive: u.IsActive,
	}
}

// Me mengembalikan data user yang sedang login, diambil dari token (via
// context yang disisipkan middleware.RequireAuth).
//
// GET /api/auth/me   (wajib header: Authorization: Bearer <token>)
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var user models.User
	if err := h.DB.Preload("Role").First(&user, userID).Error; err != nil {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toUserProfile(user))
}
