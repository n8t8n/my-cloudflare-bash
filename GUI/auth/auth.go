package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

const (
	CookieName       = "cf-session"
	CookieValue      = "cf-admin"
	CookieMaxAge     = 3600
	EnvFilePath      = ".env"
	PasswordFilePath = "password.dat"
)

var (
	hashedPassword []byte
	passwordMutex  sync.RWMutex
	devMode        bool
)

// SetDevMode sets the development mode
func SetDevMode(isDev bool) {
	devMode = isDev
}

func init() {
	// Load the hashed password from the .dat file on startup
	loadPasswordFromFile()
}

func loadPasswordFromFile() {
	passwordMutex.Lock()
	defer passwordMutex.Unlock()

	// Debugging: Print the file path
	fmt.Println("Password File Path:", PasswordFilePath)

	// Read the hashed password from the .dat file
	if _, err := os.Stat(PasswordFilePath); err == nil {
		hashedPassword, err = os.ReadFile(PasswordFilePath)
		if err != nil {
			panic("Failed to read password file: " + err.Error())
		}
		// Debugging: Print the file contents
		fmt.Println("File Contents:", string(hashedPassword))
	} else {
		// If the file doesn't exist, hash the default password and save it
		var err error
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			panic("Failed to hash password: " + err.Error())
		}
		if err := savePasswordToFile(hashedPassword); err != nil {
			panic("Failed to save password file: " + err.Error())
		}
	}
}

func savePasswordToFile(password []byte) error {
	return os.WriteFile(PasswordFilePath, password, 0644)
}

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool `json:"success"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type ChangePasswordResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func SetSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    CookieValue,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteLaxMode,
		Domain:   "",
	})
}

func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteLaxMode,
	})
}

func IsAuthenticated(r *http.Request) bool {
	c, err := r.Cookie(CookieName)
	return err == nil && c.Value == CookieValue
}

func ValidatePassword(password string) bool {
	passwordMutex.RLock()
	defer passwordMutex.RUnlock()
	fmt.Println("Hashed Password:", string(hashedPassword)) // Debugging line
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	return err == nil
}

func ChangePassword(oldPassword, newPassword string) bool {
	passwordMutex.Lock()
	defer passwordMutex.Unlock()

	// Validate the old password
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(oldPassword))
	if err != nil {
		return false
	}

	// Hash the new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false
	}

	// Update the hashed password
	hashedPassword = newHashedPassword

	// Save the new hashed password to the .dat file
	if err := savePasswordToFile(newHashedPassword); err != nil {
		return false
	}

	// Update the .env file
	if err := updateEnvFile(newPassword); err != nil {
		return false
	}

	return true
}

func updateEnvFile(newPassword string) error {
	// Read the existing .env file
	file, err := os.ReadFile(EnvFilePath)
	if err != nil {
		return err
	}

	// Split the file into lines
	lines := strings.Split(string(file), "\n")

	// Find and update the password line
	updated := false
	for i, line := range lines {
		if strings.HasPrefix(line, "PASSWORD=") {
			lines[i] = "PASSWORD=" + newPassword
			updated = true
			break
		}
	}

	// If the password line was not found, add it
	if !updated {
		lines = append(lines, "PASSWORD="+newPassword)
	}

	// Write the updated lines back to the .env file
	return os.WriteFile(EnvFilePath, []byte(strings.Join(lines, "\n")), 0644)
}
