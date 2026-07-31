package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"

	"be-eventgate/internal/auth"
	"be-eventgate/internal/httpx"
	"be-eventgate/internal/models"
)

type AuthHandler struct {
	DB           *gorm.DB
	JWTSecret    string
	JWTExpiryHrs int
}

func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpiryHrs int) *AuthHandler {
	return &AuthHandler{DB: db, JWTSecret: jwtSecret, JWTExpiryHrs: jwtExpiryHrs}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// Login memproses autentikasi pengguna dengan memvalidasi kredensial.
// Apabila berhasil, sistem akan mengembalikan JSON Web Token (JWT).
// Rute: POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var user models.User
	err := h.DB.Preload("Role").Where("email = ?", req.Email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Menggunakan pesan kesalahan generik demi keamanan guna
		// mencegah eksploitasi pencacahan akun (account enumeration).
		httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to process login")
		return
	}

	if !user.IsActive {
		httpx.WriteError(w, http.StatusForbidden, "account is inactive, please contact administrator")
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateToken(h.JWTSecret, h.JWTExpiryHrs, user.ID, user.Username, user.Role.RoleName)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	now := time.Now()
	h.DB.Model(&user).Update("last_login_at", now)

	httpx.WriteJSON(w, http.StatusOK, LoginResponse{
		Token: token,
		User:  toUserProfile(user),
	})
}
