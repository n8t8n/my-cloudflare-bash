package auth

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

const (
	CookieName   = "cf-session"
	CookieValue  = "cf-admin"
	CookieMaxAge = 3600
)

// Hashed password for "cf-manager"
var hashedPassword = []byte("$2a$10$3eZ7Xz8X1Xz8X1Xz8X1XzO.3eZ7Xz8X1Xz8X1Xz8X1XzO")

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool `json:"success"`
}

func SetSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    CookieValue,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		HttpOnly: true,
		Secure:   false,
	})
}

func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
	})
}

func IsAuthenticated(r *http.Request) bool {
	c, err := r.Cookie(CookieName)
	return err == nil && c.Value == CookieValue
}

func ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	return err == nil
}

// HashPassword generates a bcrypt hash for the given password
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
