package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"cf-manager/auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			log.Printf("Authentication failed for %s: session cookie not found", r.RemoteAddr)
		} else {
			log.Printf("Session cookie found for %s: %s", r.RemoteAddr, cookie.Value)
		}

		if !auth.IsAuthenticated(r) {
			if r.Header.Get("Content-Type") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
