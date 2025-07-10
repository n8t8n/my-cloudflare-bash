package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cf-manager/auth"
	"cf-manager/handlers"
	"cf-manager/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(1*time.Minute), 20) // 20 requests per minute

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply rate limiting only to the login endpoint
		if r.URL.Path == "/login" && r.Method == "POST" {
			if !limiter.Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the origin from the request
		origin := r.Header.Get("Origin")

		// Allow requests from localhost and local network IPs
		if origin != "" {
			// Allow localhost and 192.168.x.x, 10.x.x.x, 172.16-31.x.x ranges
			if strings.Contains(origin, "localhost") ||
				strings.Contains(origin, "127.0.0.1") ||
				strings.Contains(origin, "192.168.") ||
				strings.Contains(origin, "10.") ||
				(strings.Contains(origin, "172.") && isPrivateIP172(origin)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else {
			// If no origin header, allow all (for direct IP access)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Helper function to check if IP is in 172.16.0.0/12 range
func isPrivateIP172(origin string) bool {
	// Simple check for 172.16-31.x.x range
	parts := strings.Split(origin, ".")
	if len(parts) >= 2 {
		if parts[0] == "172" {
			// Check if second octet is between 16-31
			secondOctet := parts[1]
			return secondOctet >= "16" && secondOctet <= "31"
		}
	}
	return false
}

func main() {
	// Set dev mode based on command-line argument or environment variable
	devMode := os.Getenv("DEV_MODE") == "true"
	auth.SetDevMode(devMode)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Validate required environment variables
	requiredVars := []string{"CF_API_TOKEN", "CF_ZONE_ID", "CF_DOMAIN"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Required environment variable %s is not set", v)
		}
	}

	r := mux.NewRouter()

	// Apply CORS middleware first
	r.Use(corsMiddleware)

	// Public routes
	r.HandleFunc("/", handlers.IndexHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")
	r.HandleFunc("/change-password", handlers.ChangePasswordHandler).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/dashboard", handlers.DashboardHandler).Methods("GET")

	// DNS Management routes
	protected.HandleFunc("/dns/records", handlers.ListDNSRecordsHandler).Methods("GET")
	protected.HandleFunc("/dns/records", handlers.CreateDNSRecordHandler).Methods("POST")
	protected.HandleFunc("/dns/records/{id}", handlers.UpdateDNSRecordHandler).Methods("PUT")
	protected.HandleFunc("/dns/records/{id}", handlers.DeleteDNSRecordHandler).Methods("DELETE")

	// Tunnel Management routes
	protected.HandleFunc("/tunnels", handlers.ListTunnelsHandler).Methods("GET")
	protected.HandleFunc("/tunnels", handlers.CreateTunnelHandler).Methods("POST")
	protected.HandleFunc("/tunnels/{name}", handlers.DeleteTunnelHandler).Methods("DELETE")
	protected.HandleFunc("/tunnels/{name}/start", handlers.StartTunnelHandler).Methods("POST")
	protected.HandleFunc("/tunnels/{name}/stop", handlers.StopTunnelHandler).Methods("POST")
	protected.HandleFunc("/tunnels/{name}/status", handlers.GetTunnelStatusHandler).Methods("GET")
	protected.HandleFunc("/tunnels/{name}/config", handlers.EditTunnelConfigHandler).Methods("GET", "PUT")

	// System routes
	protected.HandleFunc("/system/status", handlers.SystemStatusHandler).Methods("GET")

	// Apply rate limiting middleware
	r.Use(rateLimitMiddleware)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Cloudflare Manager running on :%s", port)
	log.Printf("Dev mode: %v", devMode)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, r))
}
